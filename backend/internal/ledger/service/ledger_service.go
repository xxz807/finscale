package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/xxz807/finscale/backend/internal/ledger/domain"
)

// PostingRequest 定义记账请求的 DTO (Input)
type PostingRequest struct {
	ReferenceID string
	TxType      string
	Description string
	Entries     []PostingEntry
}

type PostingEntry struct {
	AccountCode string
	Direction   string // "D" or "C"
	Amount      string // 传字符串防止精度丢失
}

// LedgerService 核心服务
type LedgerService struct {
	db          *gorm.DB // 用于开启事务
	accountRepo domain.AccountRepository
	txRepo      domain.TransactionRepository
}

func NewLedgerService(db *gorm.DB, accRepo domain.AccountRepository, txRepo domain.TransactionRepository) *LedgerService {
	return &LedgerService{
		db:          db,
		accountRepo: accRepo,
		txRepo:      txRepo,
	}
}

// PostTransaction 执行记账 (ACID Transaction Script)
func (s *LedgerService) PostTransaction(ctx context.Context, req PostingRequest) (*domain.Transaction, error) {
	// 1. 基础校验
	if len(req.Entries) < 2 {
		return nil, errors.New("transaction must have at least 2 postings")
	}

	// 2. 幂等性检查 (快速失败)
	exists, err := s.txRepo.ExistsByRefID(ctx, req.ReferenceID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("transaction %s already exists", req.ReferenceID)
	}

	// 3. 准备数据 & 试算平衡检查 (In-Memory Check)
	var totalDebit, totalCredit decimal.Decimal
	var postings []domain.Posting

	// 缓存查出来的账户，避免后面更新时重复查
	accountMap := make(map[string]*domain.Account)

	for _, entry := range req.Entries {
		// 解析金额
		amt, err := decimal.NewFromString(entry.Amount)
		if err != nil {
			return nil, fmt.Errorf("invalid amount format: %s", entry.Amount)
		}
		if amt.LessThanOrEqual(decimal.Zero) {
			return nil, errors.New("amount must be positive")
		}

		// 累加借贷
		if entry.Direction == string(domain.Debit) {
			totalDebit = totalDebit.Add(amt)
		} else if entry.Direction == string(domain.Credit) {
			totalCredit = totalCredit.Add(amt)
		} else {
			return nil, errors.New("invalid direction")
		}

		// 预加载账户信息
		acc, err := s.accountRepo.FindByCode(ctx, entry.AccountCode)
		if err != nil {
			return nil, fmt.Errorf("account not found: %s", entry.AccountCode)
		}
		accountMap[entry.AccountCode] = acc

		postings = append(postings, domain.Posting{
			AccountID: acc.ID,
			Direction: domain.Direction(entry.Direction),
			Amount:    amt,
		})
	}

	// 核心逻辑：借贷必相等
	if !totalDebit.Equal(totalCredit) {
		return nil, fmt.Errorf("imbalance: debit=%s, credit=%s", totalDebit, totalCredit)
	}

	// 4. 开启数据库事务 (The Big Transaction)
	txEntity := &domain.Transaction{
		ReferenceID: req.ReferenceID,
		TxType:      req.TxType,
		Description: req.Description,
		PostedAt:    time.Now(),
		Postings:    postings,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// A. 保存交易主表和分录 (Insert Logs)
		if err := s.txRepo.Create(ctx, tx, txEntity); err != nil {
			return err
		}

		// B. 更新余额 (Update States)
		for _, entry := range req.Entries {
			acc := accountMap[entry.AccountCode]
			changeAmount := decimal.RequireFromString(entry.Amount) // 已校验过

			// 计算数据库里的增减值 (Signed Amount)
			// 资产类：借是+，贷是-
			// 负债类：贷是+，借是-
			var dbDelta decimal.Decimal

			// 获取账户类型的正负逻辑
			// 这里复用了我们之前在 domain/types.go 里定义的 AccountType 逻辑
			// 假设你已经在 domain 里加上了 Helper 方法，如果没有，这里手动判断：
			multiplier := 1
			if acc.Type == domain.Asset || acc.Type == domain.Expense {
				if entry.Direction == string(domain.Credit) {
					multiplier = -1
				}
			} else {
				// 负债、权益、收入
				if entry.Direction == string(domain.Debit) {
					multiplier = -1
				}
			}

			dbDelta = changeAmount.Mul(decimal.NewFromInt(int64(multiplier)))

			// 执行乐观锁更新
			// 注意：这里传的是 dbDelta 的字符串，比如 "-100.00"
			if err := s.accountRepo.UpdateBalance(ctx, tx, acc.ID, dbDelta.String(), acc.Version); err != nil {
				// 如果失败，Transaction 函数会自动回滚整个事务
				return fmt.Errorf("failed to update account %s: %w", acc.AccountCode, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return txEntity, nil
}
