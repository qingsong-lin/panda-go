package log

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"panda-go/common/utils"
	"panda-go/common/utils/desensitization"
)

var logger *zap.Logger

var level *zap.AtomicLevel

const (
	LOG_LEVEL_DEBUG = 1
	LOG_LEVEL_INFO  = 2
	LOG_LEVEL_WARN  = 3
	LOG_LEVEL_ERROR = 4
	LOG_LEVEL_PANIC = 5
	LOG_LEVEL_FATAL = 6
)

var logFuncMap = map[int]func(*zap.Logger, string, ...zap.Field){
	LOG_LEVEL_DEBUG: (*zap.Logger).Debug,
	LOG_LEVEL_INFO:  (*zap.Logger).Info,
	LOG_LEVEL_WARN:  (*zap.Logger).Warn,
	LOG_LEVEL_ERROR: (*zap.Logger).Error,
	LOG_LEVEL_PANIC: (*zap.Logger).Panic,
	LOG_LEVEL_FATAL: (*zap.Logger).Fatal,
}

var sugaredLogFuncMap = map[int]func(sugaredLogger *zap.SugaredLogger, template string, args ...interface{}){
	LOG_LEVEL_DEBUG: (*zap.SugaredLogger).Debugf,
	LOG_LEVEL_INFO:  (*zap.SugaredLogger).Infof,
	LOG_LEVEL_WARN:  (*zap.SugaredLogger).Warnf,
	LOG_LEVEL_ERROR: (*zap.SugaredLogger).Errorf,
	LOG_LEVEL_PANIC: (*zap.SugaredLogger).Panicf,
	LOG_LEVEL_FATAL: (*zap.SugaredLogger).Fatalf,
}

var lowercaseStr2LevelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func init() {
	config := zap.NewProductionConfig()
	//config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	level = &(config.Level)
	logger, _ = config.Build()
	logger = logger.WithOptions(zap.AddCallerSkip(2))
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.LstdFlags)
}

func GetLogger() *zap.Logger {
	return logger
}

func SetLevel(levelStr string) error {
	if l, ok := lowercaseStr2LevelMap[levelStr]; ok {
		level.SetLevel(l)
		logger.Info("[logger] change level to ", zap.String("level", levelStr), zap.Int8("level code", int8(l)))
		return nil
	} else {
		logger.Info("[logger] change level fail", zap.String("level", levelStr))
		level.SetLevel(zapcore.InfoLevel)
		return errors.New("[logger] change level fail")
	}
}

func extractCtxInfo(ctx context.Context, fields ...zap.Field) []zap.Field {
	if userId := ctx.Value("user_id"); userId != nil {
		fields = append(fields, zap.Any("user_id", userId))
	}
	if traceId := utils.GetTraceIdFromCtx(ctx); traceId != "" {
		fields = append(fields, zap.String("trace_id", traceId))
	}
	return fields
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	doLog(ctx, LOG_LEVEL_INFO, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	doLog(ctx, LOG_LEVEL_WARN, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	doLog(ctx, LOG_LEVEL_ERROR, msg, fields...)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	doLog(ctx, LOG_LEVEL_DEBUG, msg, fields...)
}

func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	doLog(ctx, LOG_LEVEL_PANIC, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	doLog(ctx, LOG_LEVEL_FATAL, msg, fields...)
}

func doLog(ctx context.Context, level int, msg string, fields ...zap.Field) {
	fields = extractCtxInfo(ctx, fields...)
	// fieldList 储存过滤完毕的值,不影响field...里的值
	fieldList := make([]zap.Field, len(fields))
	for i, one := range fields {
		fieldList[i] = *desensitization.FilterField(&one)
	}

	// 传入过滤完毕的值去实际的log函数记录日志
	logFuncMap[level](logger, msg, fieldList...)
}

func Infof(ctx context.Context, template string, args ...interface{}) {
	doSugaredLog(ctx, LOG_LEVEL_INFO, template, args...)
}

func Warnf(ctx context.Context, template string, args ...interface{}) {
	doSugaredLog(ctx, LOG_LEVEL_WARN, template, args...)
}

func Errorf(ctx context.Context, template string, args ...interface{}) {
	doSugaredLog(ctx, LOG_LEVEL_ERROR, template, args...)
}

func Debugf(ctx context.Context, template string, args ...interface{}) {
	doSugaredLog(ctx, LOG_LEVEL_DEBUG, template, args...)
}

func Panicf(ctx context.Context, template string, args ...interface{}) {
	doSugaredLog(ctx, LOG_LEVEL_PANIC, template, args...)
}

func Fatalf(ctx context.Context, template string, args ...interface{}) {
	doSugaredLog(ctx, LOG_LEVEL_FATAL, template, args...)
}

func doSugaredLog(ctx context.Context, level int, template string, args ...interface{}) {
	sugaredLogFuncMap[level](logger.Sugar(), template, args...)
}
