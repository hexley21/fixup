package zap

import (
	"os"

	"github.com/hexley21/handy/pkg/config"
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


func InitLogger(cfg config.Logging, isProduction bool) *zapLogger {
	logWriter := zapcore.AddSync(os.Stdout)

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

	core := zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(getLoggerLevel(cfg.LogLevel)))

	var options []zap.Option

	if cfg.CallerEnabled {
		options = append(options, zap.AddCaller())
		options = append(options, zap.AddCallerSkip(1))
	}

	return &zapLogger{sugarLogger: zap.New(core, options...).Sugar()}
}

func (l *zapLogger) Debug(msg string, args ...any) {
	l.sugarLogger.Debugw(msg, args...)
}

func (l *zapLogger) Info(msg string, args ...any) {
	l.sugarLogger.Infow(msg, args...)
}

func (l *zapLogger) Warn(msg string, args ...any) {
	l.sugarLogger.Warnw(msg, args...)
}

func (l *zapLogger) Error(err error, args ...any) {
	l.sugarLogger.Errorw(err.Error(), args...)
}

func (l *zapLogger) ErrorCause(err error, cause any) {
	l.Error(err, "cause", cause)
}


func (l *zapLogger) Fatal(err error, args ...any) {
	l.sugarLogger.Fatalw(err.Error(), args...)
}
