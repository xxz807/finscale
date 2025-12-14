package repo

import (
	"context"
	"errors"

	"github.com/xxz807/finscale/backend/internal/ledger/domain"
	"gorm.io/gorm"
)

type PostgresAccountRepo struct {
	db *gorm.DB
}

func NewAccountRepo(db *gorm.DB) *PostgresAccountRepo {
	return &PostgresAccountRepo{db: db}
}

func (r *PostgresAccountRepo) FindByCode(ctx context.Context, code string) (*domain.Account, error) {
	var account domain.Account
	// 注意：这里我们不做 Select For Update，因为我们要演示乐观锁
	if err := r.db.WithContext(ctx).Where("account_code = ?", code).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *PostgresAccountRepo) FindByID(ctx context.Context, id int64) (*domain.Account, error) {
	var account domain.Account
	if err := r.db.WithContext(ctx).First(&account, id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// UpdateBalance 实现乐观锁更新
// SQL: UPDATE accounts SET balance = balance + ?, version = version + 1 WHERE id = ? AND version = ?
func (r *PostgresAccountRepo) UpdateBalance(ctx context.Context, tx *gorm.DB, id int64, amount string, version int64) error {
	// 注意：必须使用传入的 tx (事务会话)，而不是 r.db

	// GORM 的 Update 语句
	result := tx.WithContext(ctx).Model(&domain.Account{}).
		Where("id = ? AND version = ?", id, version).
		Updates(map[string]interface{}{
			"balance": gorm.Expr("balance + ?", amount), // 数据库层面的加减，更安全
			"version": gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return result.Error
	}

	// 关键点：如果没有行被更新，说明 version 不匹配（被别人改过了）
	if result.RowsAffected == 0 {
		return errors.New("optimistic lock conflict: account modified by others")
	}

	return nil
}

// ---------------------------------------------------------

type PostgresTransactionRepo struct {
	db *gorm.DB
}

func NewTransactionRepo(db *gorm.DB) *PostgresTransactionRepo {
	return &PostgresTransactionRepo{db: db}
}

func (r *PostgresTransactionRepo) ExistsByRefID(ctx context.Context, refID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Transaction{}).Where("reference_id = ?", refID).Count(&count).Error
	return count > 0, err
}

func (r *PostgresTransactionRepo) Create(ctx context.Context, tx *gorm.DB, t *domain.Transaction) error {
	// GORM 会自动处理 Transaction -> Postings 的关联插入
	return tx.WithContext(ctx).Create(t).Error
}
