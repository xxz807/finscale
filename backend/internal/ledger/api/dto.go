package api

// PostTransactionReq 对应前端发来的 JSON
type PostTransactionReq struct {
	ReferenceID string       `json:"reference_id" binding:"required"`
	TxType      string       `json:"tx_type" binding:"required"`
	Description string       `json:"description"`
	Postings    []PostingReq `json:"postings" binding:"required,min=2"` // 至少要有借和贷两条
}

type PostingReq struct {
	AccountCode string `json:"account_code" binding:"required"`
	Direction   string `json:"direction" binding:"required,oneof=D C"` // 只能是 D 或 C
	Amount      string `json:"amount" binding:"required"`              // 必须传字符串
}
