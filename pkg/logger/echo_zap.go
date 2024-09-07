package logger

import (
	"io"
	"os"

	"github.com/hexley21/fixup/pkg/config"
	"github.com/labstack/gommon/log"
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

func NewZapLogger(cfg config.Logging, isProduction bool) *zapLogger {
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

func (l *zapLogger) Output() io.Writer {
	return os.Stderr
}

func (l *zapLogger) SetOutput(_ io.Writer) {}

func (l *zapLogger) Prefix() string {
	return ""
}

func (l *zapLogger) SetPrefix(p string) {
	l.sugarLogger = l.sugarLogger.Named(p)
}

func (l *zapLogger) Level() log.Lvl {
	switch l.sugarLogger.Level() {
	case zap.DebugLevel:
		return log.DEBUG
	case zap.InfoLevel:
		return log.INFO
	case zap.WarnLevel:
		return log.WARN
	default:
		return log.ERROR
	}
}

func (l *zapLogger) SetLevel(_ log.Lvl) {}

func (l *zapLogger) SetHeader(_ string) {}

func (l *zapLogger) Print(i ...any) {
	l.sugarLogger.Info(i...)
}

func (l *zapLogger) Printf(format string, args ...any) {
	l.sugarLogger.Infof(format, args...)
}

func (l *zapLogger) Printj(j log.JSON) {
	var args []any

	for k, v := range j {
		args = append(args, k, v)
	}

	l.sugarLogger.Info(args...)
}

func (l *zapLogger) Debug(i ...any) {
	l.sugarLogger.Debug(i...)
}

func (l *zapLogger) Debugf(format string, args ...any) {
	l.sugarLogger.Debugf(format, args...)
}

func (l *zapLogger) Debugj(j log.JSON) {
	var args []any

	for k, v := range j {
		args = append(args, k, v)
	}

	l.sugarLogger.Debugw("json", args...)
}

func (l *zapLogger) Info(i ...any) {
	l.sugarLogger.Info(i...)
}

func (l *zapLogger) Infof(format string, args ...any) {
	l.sugarLogger.Infof(format, args...)
}

func (l *zapLogger) Infoj(j log.JSON) {
	var args []any

	for k, v := range j {
		args = append(args, k, v)
	}

	l.sugarLogger.Infow("json", args...)
}

func (l *zapLogger) Warn(i ...any) {
	l.sugarLogger.Warn(i...)
}

func (l *zapLogger) Warnf(format string, args ...any) {
	l.sugarLogger.Warnf(format, args...)
}

func (l *zapLogger) Warnj(j log.JSON) {
	var args []any

	for k, v := range j {
		args = append(args, k, v)
	}

	l.sugarLogger.Warnw("json", args...)
}

func (l *zapLogger) Error(i ...any) {
	l.sugarLogger.Error(i...)
}

func (l *zapLogger) Errorf(format string, args ...any) {
	l.sugarLogger.Errorf(format, args...)
}

func (l *zapLogger) Errorj(j log.JSON) {
	var args []any

	for k, v := range j {
		args = append(args, k, v)
	}

	l.sugarLogger.Errorw("json", args...)
}

func (l *zapLogger) Fatal(i ...any) {
	l.sugarLogger.Fatal(i...)
}

func (l *zapLogger) Fatalj(j log.JSON) {
	var args []any

	for k, v := range j {
		args = append(args, k, v)
	}

	l.sugarLogger.Fatalw("json", args...)
}

func (l *zapLogger) Fatalf(format string, args ...any) {
	l.sugarLogger.Fatalf(format, args...)
}

func (l *zapLogger) Panic(i ...any) {
	l.sugarLogger.Panic(i...)
}

func (l *zapLogger) Panicj(j log.JSON) {
	var args []any

	for k, v := range j {
		args = append(args, k, v)
	}

	l.sugarLogger.Panicw("json", args...)
}

func (l *zapLogger) Panicf(format string, args ...any) {
	l.sugarLogger.Panicf(format, args...)
}
