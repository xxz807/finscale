package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xxz807/finscale/backend/internal/ledger/api"
)

// Server 封装 HTTP 服务
type Server struct {
	engine *gin.Engine
	logger *zap.Logger
	port   string
	server *http.Server
}

// NewServer 初始化 HTTP Server (包含网关逻辑)
func NewServer(
	logger *zap.Logger,
	cfgPort string,
	cfgMode string,
	// 依赖注入：传入具体的 Handler
	ledgerHandler *api.LedgerHandler,
) *Server {

	// 1. 设置 Gin 模式
	if cfgMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// ==========================================
	// 🏗️ Logical Gateway Layer (逻辑网关层)
	// ==========================================

	// 1. Recovery (防崩)
	r.Use(gin.Recovery())

	// 2. Custom Logger (接入 Zap)
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next() // 执行后续逻辑

		cost := time.Since(start)
		logger.Info("HTTP Request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Duration("cost", cost),
		)
	})

	// 3. CORS (跨域处理 - 允许前端访问)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 4. Dummy Auth (模拟鉴权 - MVP阶段)
	// 以后这里会替换成真正的 JWT 中间件
	r.Use(func(c *gin.Context) {
		// 假装从 Token 解析出了 UserID
		c.Set("x-user-id", "admin-001")
		c.Next()
	})

	// ==========================================
	// 🚦 Routing Layer (路由分发)
	// ==========================================

	v1 := r.Group("/api/v1")
	{
		// 注册 Ledger 模块的路由
		ledgerHandler.RegisterRoutes(v1)

		// 健康检查
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "UP"})
		})
	}

	return &Server{
		engine: r,
		logger: logger,
		port:   cfgPort,
	}
}

// Run 启动服务
func (s *Server) Run() error {
	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: s.engine,
	}
	s.logger.Info("🚀 FinScale Logical Gateway started", zap.String("port", s.port))
	return s.server.ListenAndServe()
}

// Shutdown 优雅停机 (Graceful Shutdown)
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
