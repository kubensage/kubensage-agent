package utils

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
)

func SetupLogger(flags Config) *zap.Logger {
	logger, err := newLogger(&flags.LogLevel, &flags.LogFile, &flags.LogMaxSize, &flags.LogMaxBackups,
		&flags.LogMaxAge, &flags.LogCompress,
	)

	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	return logger
}

func newLogger(logLevel *string, file *string, size *int, backups *int, age *int, compress *bool) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := (&level).UnmarshalText([]byte(*logLevel)); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   *file,
		MaxSize:    *size,
		MaxBackups: *backups,
		MaxAge:     *age,
		Compress:   *compress,
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
