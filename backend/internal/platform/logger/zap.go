package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger 初始化 Zap Logger
// 根据环境不同（Dev/Prod）输出不同格式
func NewLogger(mode string) *zap.Logger {
	var config zap.Config

	if mode == "debug" {
		// 开发环境：控制台彩色输出，人类可读
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		// 生产环境：JSON 输出，机器可读 (ELK)
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := config.Build()
	if err != nil {
		os.Exit(1)
	}

	return logger
}
