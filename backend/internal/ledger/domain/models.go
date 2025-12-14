package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Account 会计科目实体
// 对应数据库表: ledger.accounts
type Account struct {
	ID          int64           `gorm:"primaryKey;autoIncrement"`
	AccountCode string          `gorm:"uniqueIndex;type:varchar(32);not null"`
	Name        string          `gorm:"type:varchar(100);not null"`
	Type        AccountType     `gorm:"type:smallint;not null"`
	Currency    string          `gorm:"type:char(3);default:'CNY';not null"`
	Balance     decimal.Decimal `gorm:"type:decimal(20,4);not null;default:0"`
	Version     int64           `gorm:"not null;default:1"` // 乐观锁
	Status      int16           `gorm:"type:smallint;default:1"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName 指定 schema 和表名
func (Account) TableName() string {
	return "ledger.accounts"
}

// Transaction 交易主表实体
// 对应数据库表: ledger.transactions
type Transaction struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	ReferenceID string    `gorm:"uniqueIndex;type:varchar(64);not null"`
	TxType      string    `gorm:"type:varchar(32);not null"`
	Description string    `gorm:"type:text"`
	PostedAt    time.Time `gorm:"not null"`
	Metadata    []byte    `gorm:"type:jsonb"` // 简化处理，直接存 []byte 或 string
	CreatedAt   time.Time

	// 关联关系 (一对多)
	Postings []Posting `gorm:"foreignKey:TransactionID"`
}

func (Transaction) TableName() string {
	return "ledger.transactions"
}

// Posting 资金分录实体
// 对应数据库表: ledger.postings
type Posting struct {
	ID            int64           `gorm:"primaryKey;autoIncrement"`
	TransactionID int64           `gorm:"not null;index"`
	AccountID     int64           `gorm:"not null;index"`
	Direction     Direction       `gorm:"type:char(1);not null"`
	Amount        decimal.Decimal `gorm:"type:decimal(20,4);not null"` // 必须 > 0
	ExchangeRate  decimal.Decimal `gorm:"type:decimal(10,6);default:1"`
}

func (Posting) TableName() string {
	return "ledger.postings"
}
