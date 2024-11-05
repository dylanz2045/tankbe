package utils

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logDir = "./logs/"
)

func Initlogger() error {
	// 检查日志目录是否存在
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		// 目录不存在，创建目录
		err := os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			fmt.Printf("无法创建目录：%v\n", err)
			return err
		}
		fmt.Println("目录已创建：", logDir)
	} else {
		fmt.Println("目录已存在：", logDir)
	}

	// 配置 lumberjack 管理日志文件
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logDir + "logs.txt", // 日志文件路径
		MaxSize:    5,                   // 每个日志文件的最大大小（单位：MB）
		MaxBackups: 10,                  // 最大保留的旧日志文件数量
		MaxAge:     30,                  // 最长保留天数（单位：天）
		Compress:   true,                // 是否启用压缩
	}

	// 创建文件和控制台的日志写入器
	fileWriteSyncer := zapcore.AddSync(lumberjackLogger)
	consoleWriteSyncer := zapcore.AddSync(os.Stdout)

	// 设置日志编码器
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.TimeKey = "timestamp" // 自定义时间字段名
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 创建文件和控制台两个核心
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, fileWriteSyncer, zap.InfoLevel),     // 输出到文件
		zapcore.NewCore(encoder, consoleWriteSyncer, zap.DebugLevel), // 输出到控制台
	)

	// 构建 logger
	Logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(Logger)
	Logger.Info("log init success")

	//初始化日志文实体
	loginit()

	return nil
}

func CheckLogger() error {
	if Logger == nil {
		return fmt.Errorf("日志实体消失了，请处理")
	}
	return nil
}
