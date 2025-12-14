package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxz807/finscale/backend/internal/ledger/service"
)

type LedgerHandler struct {
	svc *service.LedgerService
}

func NewLedgerHandler(svc *service.LedgerService) *LedgerHandler {
	return &LedgerHandler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *LedgerHandler) RegisterRoutes(r *gin.RouterGroup) {
	ledgerGroup := r.Group("/ledger")
	{
		ledgerGroup.POST("/transactions", h.PostTransaction)
		// 未来可以在这里加 GET /accounts/:code 查询余额
	}
}

// PostTransaction 记账接口
// POST /api/v1/ledger/transactions
func (h *LedgerHandler) PostTransaction(c *gin.Context) {
	var req PostTransactionReq

	// 1. 参数绑定与基础校验
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 2. DTO 转换 (API Layer -> Service Layer)
	// 虽然结构很像，但在架构上需要转换，解耦层级
	svcReq := service.PostingRequest{
		ReferenceID: req.ReferenceID,
		TxType:      req.TxType,
		Description: req.Description,
		Entries:     make([]service.PostingEntry, len(req.Postings)),
	}

	for i, p := range req.Postings {
		svcReq.Entries[i] = service.PostingEntry{
			AccountCode: p.AccountCode,
			Direction:   p.Direction,
			Amount:      p.Amount,
		}
	}

	// 3. 调用业务逻辑
	tx, err := h.svc.PostTransaction(c.Request.Context(), svcReq)
	if err != nil {
		// 简单的错误处理策略
		// 生产环境应该根据 err 类型判断返回 409 (Conflict) 还是 422 (Unprocessable)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message":      "Transaction posted successfully",
		"tx_id":        tx.ID,
		"reference_id": tx.ReferenceID,
		"posted_at":    tx.PostedAt,
	})
}
