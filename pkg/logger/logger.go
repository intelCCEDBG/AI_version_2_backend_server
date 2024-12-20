package logger

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

func InitLogger(path string) {
	rotateCfg := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path,
		MaxSize:    5, // megabytes
		MaxBackups: 30,
		MaxAge:     30, // days
	})
	rotatedFileWriteSyncer := zapcore.AddSync(rotateCfg)
	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)
	// console logger
	stdout := zapcore.AddSync(os.Stdout)
	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	// set core
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, zapcore.DebugLevel),
		zapcore.NewCore(fileEncoder, rotatedFileWriteSyncer, zapcore.DebugLevel),
	)
	Log = zap.New(core)
}

func Info(message string, fields ...zap.Field) {
	if Log == nil {
		return
	}
	callerFields := GetCallerInfoForLog()
	fields = append(fields, callerFields...)
	Log.Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	if Log == nil {
		return
	}
	callerFields := GetCallerInfoForLog()
	fields = append(fields, callerFields...)
	Log.Debug(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	if Log == nil {
		return
	}
	callerFields := GetCallerInfoForLog()
	fields = append(fields, callerFields...)
	Log.Error(message, fields...)
}

func Warn(message string, fields ...zap.Field) {
	if Log == nil {
		return
	}
	callerFields := GetCallerInfoForLog()
	fields = append(fields, callerFields...)
	Log.Warn(message, fields...)
}

func Panic(message string, fields ...zap.Field) {
	if Log == nil {
		return
	}
	callerFields := GetCallerInfoForLog()
	fields = append(fields, callerFields...)
	Log.Panic(message, fields...)
}

func GetCallerInfoForLog() (callerFields []zap.Field) {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	funcName = path.Base(funcName)

	callerFields = append(callerFields, zap.String("func", funcName), zap.String("file", file), zap.Int("line", line))
	return
}

func GinLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		requestURI := c.Request.RequestURI
		ginInfo := fmt.Sprintf("[GIN] %v | %3d | %13v | %15s | %-7s %s", endTime.Format("2006/01/02 - 15:04:05"), statusCode, latencyTime, clientIP, method, requestURI)
		Info(ginInfo)
	}
}
