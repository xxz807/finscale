package domain

import (
	"context"

	"gorm.io/gorm"
)

// AccountRepository 定义账户仓储接口
// 这是一个 Port (端口)，Adapter (适配器) 将在基础设施层实现它
type AccountRepository interface {
	// FindByID 根据ID查询账户
	FindByID(ctx context.Context, id int64) (*Account, error)

	// FindByCode 根据业务代码查询账户 (用于记账时查找)
	FindByCode(ctx context.Context, code string) (*Account, error)

	// UpdateBalance 核心：更新余额 (带乐观锁版本号)
	// 返回 error 包含是否并发冲突
	UpdateBalance(ctx context.Context, db *gorm.DB, id int64, amount string, version int64) error
}

// TransactionRepository 定义交易仓储接口
type TransactionRepository interface {
	// Create 保存交易主表和分录 (在一个事务中)
	Create(ctx context.Context, db *gorm.DB, tx *Transaction) error

	// ExistsByRefID 幂等性检查
	ExistsByRefID(ctx context.Context, refID string) (bool, error)
}
