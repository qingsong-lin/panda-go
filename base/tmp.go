package main

//
//import (
//	"fmt"
//	"go.uber.org/zap"
//	"go.uber.org/zap/zapcore"
//	"sync"
//	"sync/atomic"
//)
//
//var loggerAtomic atomicLogger
//
//type atomicLogger struct {
//	mu    sync.RWMutex
//	level atomic.Value
//}
//
//func (al *atomicLogger) SetLevel(level zapcore.Level) {
//	al.mu.Lock()
//	defer al.mu.Unlock()
//	al.level.Store(level)
//}
//
//func (al *atomicLogger) GetLogger() *zap.Logger {
//	al.mu.RLock()
//	defer al.mu.RUnlock()
//	currentLevel := al.level.Load()
//	if currentLevel == nil {
//		return nil
//	}
//	return newLogger(currentLevel.(zapcore.Level))
//}
//
//func newLogger(level zapcore.Level) *zap.Logger {
//	config := zap.NewDevelopmentConfig()
//	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
//	config.Level = zap.NewAtomicLevelAt(level)
//
//	logger, _ := config.Build()
//	return logger
//}
//
//func main23() {
//	// 初始化日志，默认为Info级别
//	loggerAtomic.SetLevel(zapcore.InfoLevel)
//	logger := loggerAtomic.GetLogger()
//
//	fmt.Println("Log at Info level:")
//	logger.Info("This is an Info log.")
//
//	// 调整日志级别为Debug
//	loggerAtomic.SetLevel(zapcore.DebugLevel)
//
//	// 获取新的Logger，级别为Debug
//	logger = loggerAtomic.GetLogger()
//
//	fmt.Println("Log at Debug level:")
//	logger.Debug("This is a Debug log.")
//}
