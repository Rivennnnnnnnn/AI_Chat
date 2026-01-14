package utils

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger() error {
	// 1. 设置写入器 (输出到控制台和文件)
	writeSyncer := getLogWriter()

	// 2. 设置编码器 (JSON 格式适合生产，Console 格式适合开发)
	encoder := getEncoder()

	// 3. 设置日志级别
	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	core := zapcore.NewCore(encoder, writeSyncer, level)

	// zap.AddCaller() 会记录调用日志的代码行号
	Log = zap.New(core, zap.AddCaller())
	return nil
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	// 时间格式化：2026-01-12 15:04:05
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 级别大写：INFO, ERROR
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./logs/app.log", // 日志文件路径
		MaxSize:    10,               // 每个文件最大 10MB
		MaxBackups: 5,                // 保留最近 5 个备份
		MaxAge:     30,               // 保留最近 30 天的日志
		Compress:   false,            // 是否压缩旧文件
	}
	// 同时输出到文件和终端控制台
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
}
