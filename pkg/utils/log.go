package utils

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger(logLevel string) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := (&level).UnmarshalText([]byte(logLevel)); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "/var/log/kubensage/kubensage-agent.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	})

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg), // o NewConsoleEncoder
		writer,
		level,
	)

	return zap.New(core), nil
}
