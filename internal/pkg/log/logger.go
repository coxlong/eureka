package log

import (
	"fmt"

	"github.com/coxlong/eureka/internal/pkg/config"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// InitLogger 初始化日志记录器。
func InitLogger(mode string, cfg *config.Logger) (*zap.Logger, error) {
	writeSyncer := getLogWriter(cfg.Filename, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge, cfg.Compress)
	atomicLevel, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	var encoderConfig zapcore.EncoderConfig
	if mode == config.Production {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else if mode == config.Development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		return nil, fmt.Errorf("unknown env: %s", mode)
	}
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		atomicLevel,
	)

	logger = zap.New(core, zap.AddCaller())
	return logger, nil
}

// getLogWriter 配置并返回一个logWriter
func getLogWriter(filename string, maxSize, maxBackups, maxAge int, compress bool) zapcore.WriteSyncer {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}
	return zapcore.AddSync(lumberjackLogger)
}

// GetLogger 返回初始化的全局日志记录器。
func GetLogger() *zap.Logger {
	return logger
}

// Debug 打印调试级别的日志。
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Info 打印信息级别的日志。
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Warn 打印警告级别的日志。
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Error 打印错误级别的日志。
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Fatal 打印致命级别的日志，并退出程序。
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}
