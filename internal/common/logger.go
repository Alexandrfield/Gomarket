package common

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Warnf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
	Debugf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

func GetComponentLogger() Logger {
	configLogConsole := zap.NewDevelopmentEncoderConfig()
	configLogConsole.ConsoleSeparator = " | "
	configLogConsole.EncodeTime = zapcore.ISO8601TimeEncoder
	configLogConsole.EncodeLevel = zapcore.CapitalColorLevelEncoder
	configLogConsole.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConsole := zapcore.NewConsoleEncoder(configLogConsole)

	level := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	stdout := zapcore.AddSync(os.Stdout)
	core := zapcore.NewTee(
		zapcore.NewCore(encoderConsole, stdout, level),
	)
	logger := zap.New(core, zap.AddCaller())
	return logger.Sugar()
}
