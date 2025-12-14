package main

import (
	"log"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/xxz807/finscale/backend/internal/ledger/adapter/repo"
	"github.com/xxz807/finscale/backend/internal/ledger/api"
	"github.com/xxz807/finscale/backend/internal/ledger/service"
	"github.com/xxz807/finscale/backend/internal/platform/database"
	"github.com/xxz807/finscale/backend/internal/platform/logger"
	"github.com/xxz807/finscale/backend/internal/platform/server"
)

func main() {
	// 1. 加载配置
	viper.SetConfigFile("../../configs/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	// 2. 初始化基础设施 (Infra)
	// Logger
	appLogger := logger.NewLogger(viper.GetString("server.mode"))
	// Database
	dsn := viper.GetString("database.dsn")
	max_idle_conns := viper.GetInt("database.max_idle_conns")
	max_open_conns := viper.GetInt("database.max_open_conns")
	db := database.NewPostgresDB(dsn, max_idle_conns, max_open_conns)

	// 3. 依赖注入 (Wiring)
	// -- Ledger Module --
	accountRepo := repo.NewAccountRepo(db)
	txRepo := repo.NewTransactionRepo(db)
	ledgerSvc := service.NewLedgerService(db, accountRepo, txRepo)
	ledgerHandler := api.NewLedgerHandler(ledgerSvc)

	// 4. 初始化 Server (Gateway)
	// 将 Handler 注入到 Server 中
	srv := server.NewServer(
		appLogger,
		viper.GetString("server.port"),
		viper.GetString("server.mode"),
		ledgerHandler,
	)

	// 5. 启动服务
	if err := srv.Run(); err != nil {
		appLogger.Fatal("Server startup failed", zap.Error(err))
	}
}
