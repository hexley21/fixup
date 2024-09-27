package zap_logger

import (
	"log"
	"os"

	"github.com/hexley21/fixup/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	sugarLogger *zap.SugaredLogger
}

var loggerLevelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"panic": zapcore.PanicLevel,
	"fatal": zapcore.FatalLevel,
}

func getLoggerLevel(lvl string) zapcore.Level {
	level, exist := loggerLevelMap[lvl]
	if !exist {
		return zapcore.DebugLevel
	}

	return level
}

func New(cfg config.Logging, isProduction bool) *zapLogger {
	logWriter := zapcore.AddSync(os.Stdout)

	logFile, err := os.OpenFile("./log/"+ cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatal(err)
    }
	
    fileWriter := zapcore.AddSync(logFile)

	var encoderCfg zapcore.EncoderConfig
	var encoder zapcore.Encoder

	if isProduction {
		encoderCfg = zap.NewProductionEncoderConfig()
		encoderCfg.NameKey = "[SERVICE]"
		encoderCfg.TimeKey = "[TIME]"
		encoderCfg.LevelKey = "[LEVEL]"
		encoderCfg.FunctionKey = "[CALLER]"
		encoderCfg.CallerKey = "[LINE]"
		encoderCfg.MessageKey = "[MESSAGE]"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
		encoderCfg.EncodeName = zapcore.FullNameEncoder
		encoderCfg.EncodeDuration = zapcore.StringDurationEncoder
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
		encoderCfg.NameKey = "[SERVICE]"
		encoderCfg.TimeKey = "[TIME]"
		encoderCfg.LevelKey = "[LEVEL]"
		encoderCfg.FunctionKey = "[CALLER]"
		encoderCfg.CallerKey = "[LINE]"
		encoderCfg.MessageKey = "[MESSAGE]"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.EncodeName = zapcore.FullNameEncoder
		encoderCfg.EncodeDuration = zapcore.StringDurationEncoder
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderCfg.EncodeCaller = zapcore.FullCallerEncoder
		encoderCfg.ConsoleSeparator = " | "
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	core := zapcore.NewTee(
        zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(getLoggerLevel(cfg.LogLevel))),
        zapcore.NewCore(encoder, fileWriter, zap.NewAtomicLevelAt(getLoggerLevel(cfg.LogLevel))),
    )

	var options []zap.Option

	if cfg.CallerEnabled {
		options = append(options, zap.AddCaller())
		options = append(options, zap.AddCallerSkip(2))
	}

	return &zapLogger{sugarLogger: zap.New(core, options...).Sugar()}
}

func (l *zapLogger) Debug(i ...any) {
	l.sugarLogger.Debug(i...)
}

func (l *zapLogger) Debugf(format string, args ...any) {
	l.sugarLogger.Debugf(format, args...)
}

func (l *zapLogger) Info(i ...any) {
	l.sugarLogger.Info(i...)
}

func (l *zapLogger) Infof(format string, args ...any) {
	l.sugarLogger.Infof(format, args...)
}

func (l *zapLogger) Warn(i ...any) {
	l.sugarLogger.Warn(i...)
}

func (l *zapLogger) Warnf(format string, args ...any) {
	l.sugarLogger.Warnf(format, args...)
}

func (l *zapLogger) Error(i ...any) {
	l.sugarLogger.Error(i...)
}

func (l *zapLogger) Errorf(format string, args ...any) {
	l.sugarLogger.Errorf(format, args...)
}

func (l *zapLogger) Fatal(i ...any) {
	l.sugarLogger.Fatal(i...)
}

func (l *zapLogger) Fatalf(format string, args ...any) {
	l.sugarLogger.Fatalf(format, args...)
}

func (l *zapLogger) Panic(i ...any) {
	l.sugarLogger.Panic(i...)
}

func (l *zapLogger) Panicf(format string, args ...any) {
	l.sugarLogger.Panicf(format, args...)
}

func (l *zapLogger) Print(i ...any) {
	l.sugarLogger.Info(i...)
}