package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB 初始化数据库连接
// 在 DDD 中，它属于 Infrastructure 层
func NewPostgresDB(dsn string, max_idle_conns int, max_open_conns int) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// 开启 SQL 日志，方便开发时观察
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get sql.DB: %v", err)
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(max_idle_conns)
	sqlDB.SetMaxOpenConns(max_open_conns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ Database connection established (Port: 5433)")
	return db
}
