package domain

// AccountType 账户类型 (1-5)
type AccountType int16

const (
	Asset     AccountType = 1 // 资产
	Liability AccountType = 2 // 负债
	Equity    AccountType = 3 // 权益
	Income    AccountType = 4 // 收入
	Expense   AccountType = 5 // 费用
)

// Direction 借贷方向 (D/C)
type Direction string

const (
	Debit  Direction = "D"
	Credit Direction = "C"
)

// IsValid 校验方向合法性
func (d Direction) IsValid() bool {
	return d == Debit || d == Credit
}
