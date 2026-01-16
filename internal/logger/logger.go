package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	wqlogger "github.com/WQGroup/logger"
	"github.com/sirupsen/logrus"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	Log           *logrus.Logger
	fileWriter    io.Writer
	rotateLogsWriter *rotatelogs.RotateLogs
	lumberjackWriter *lumberjack.Logger
)

// Init 初始化日志系统
func Init(cfg Config) error {
	settings := wqlogger.NewSettings()

	// 设置日志级别
	settings.Level = parseLevel(cfg.Level)

	// 设置日志文件路径和名称
	logName := "server"
	logDir := "./logs"
	if cfg.FilePath != "" {
		// 从完整路径中提取目录和文件名
		// 例如: "./logs/server.log" -> dir: "./logs", name: "server"
		dir := filepath.Dir(cfg.FilePath)
		filename := filepath.Base(cfg.FilePath)

		// 如果目录不是空的，使用该目录
		if dir != "." && dir != "" {
			logDir = dir
		}

		// 去除扩展名
		if idx := strings.LastIndex(filename, "."); idx > 0 {
			logName = filename[:idx]
		} else {
			logName = filename
		}
	}

	// WQGroup/logger 会自动在 LogRootFPath 下创建 Logs 目录
	settings.LogRootFPath = logDir
	settings.LogNameBase = logName

	// 设置文件大小限制（转换为 MB）
	if cfg.MaxSize > 0 {
		settings.MaxSizeMB = int(cfg.MaxSize / 1024 / 1024)
	}

	// 设置保留天数
	if cfg.MaxAge > 0 {
		settings.MaxAgeDays = cfg.MaxAge
	}

	// 使用配置初始化日志
	wqlogger.SetLoggerSettings(settings)

	// 获取日志实例
	var err error
	Log, err = wqlogger.GetLogger()
	if err != nil {
		return err
	}

	// 创建文件 writer 以支持不同的输出模式
	if err := createFileWriter(settings, cfg.FilePath); err != nil {
		return err
	}

	// 根据 Output 配置设置输出目标
	applyOutputMode(cfg.Output)

	return nil
}

// createFileWriter 创建文件写入器
func createFileWriter(settings *wqlogger.Settings, logPath string) error {
	pathRoot := filepath.Join(settings.LogRootFPath, "Logs")
	if settings.LogRootFPath != "." {
		pathRoot = settings.LogRootFPath
	}
	if _, err := os.Stat(pathRoot); os.IsNotExist(err) {
		err = os.MkdirAll(pathRoot, 0750)
		if err != nil {
			return err
		}
	}

	var err error
	if settings.MaxSizeMB > 0 {
		// 大小轮转模式
		logDir := pathRoot
		if _, err = os.Stat(logDir); os.IsNotExist(err) {
			err = os.MkdirAll(logDir, 0750)
			if err != nil {
				return err
			}
		}

		logFilePath := filepath.Join(logDir, settings.LogNameBase+".log")
		lumberjackWriter = &lumberjack.Logger{
			Filename:  logFilePath,
			MaxSize:   settings.MaxSizeMB,
			MaxAge:    settings.MaxAgeDays,
			LocalTime: true,
			Compress:  false,
		}
		fileWriter = lumberjackWriter
		rotateLogsWriter = nil
	} else {
		// 时间轮转模式
		logPattern := filepath.Join(pathRoot, settings.LogNameBase+"--%Y%m%d%H%M--.log")

		rotateLogsWriter, err = rotatelogs.New(
			logPattern,
			rotatelogs.WithMaxAge(settings.MaxAge),
			rotatelogs.WithRotationTime(settings.RotationTime),
		)
		if err != nil {
			return err
		}
		fileWriter = rotateLogsWriter
		lumberjackWriter = nil
	}

	return nil
}

// applyOutputMode 根据 output 配置设置日志输出目标
func applyOutputMode(output string) {
	if Log == nil {
		return
	}

	switch strings.ToLower(output) {
	case "console":
		// 仅控制台输出
		Log.SetOutput(os.Stderr)
	case "file":
		// 仅文件输出
		if fileWriter != nil {
			Log.SetOutput(fileWriter)
		}
	case "both":
		// 同时输出到控制台和文件
		if fileWriter != nil {
			Log.SetOutput(io.MultiWriter(os.Stderr, fileWriter))
		}
	default:
		// 默认行为：both
		if fileWriter != nil {
			Log.SetOutput(io.MultiWriter(os.Stderr, fileWriter))
		}
	}
}

// Config 日志配置结构体
type Config struct {
	Level      string `yaml:"level"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"filePath"`
	MaxSize    int64  `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
	Compress   bool   `yaml:"compress"`
}

// parseLevel 解析日志级别字符串
func parseLevel(level string) logrus.Level {
	switch strings.ToLower(level) {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

// 包级别的便捷函数，直接委托给 WQGroup/logger

func Debug(args ...interface{}) {
	if Log != nil {
		Log.Debug(args...)
	}
}

func Info(args ...interface{}) {
	if Log != nil {
		Log.Info(args...)
	}
}

func Warn(args ...interface{}) {
	if Log != nil {
		Log.Warn(args...)
	}
}

func Error(args ...interface{}) {
	if Log != nil {
		Log.Error(args...)
	}
}

func Fatal(args ...interface{}) {
	if Log != nil {
		Log.Fatal(args...)
	}
}

func Panic(args ...interface{}) {
	if Log != nil {
		Log.Panic(args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if Log != nil {
		Log.Debugf(format, args...)
	}
}

func Infof(format string, args ...interface{}) {
	if Log != nil {
		Log.Infof(format, args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if Log != nil {
		Log.Warnf(format, args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if Log != nil {
		Log.Errorf(format, args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	if Log != nil {
		Log.Fatalf(format, args...)
	}
}

func Panicf(format string, args ...interface{}) {
	if Log != nil {
		Log.Panicf(format, args...)
	}
}

// WithField 返回一个带字段的日志条目
func WithField(key string, value interface{}) *logrus.Entry {
	if Log != nil {
		return Log.WithField(key, value)
	}
	return logrus.NewEntry(logrus.New())
}

// WithFields 返回一个带多个字段的日志条目
func WithFields(fields map[string]interface{}) *logrus.Entry {
	if Log != nil {
		return Log.WithFields(fields)
	}
	return logrus.NewEntry(logrus.New())
}

// Close 关闭日志系统
func Close() error {
	// 关闭 lumberjack writer（如果存在）
	if lumberjackWriter != nil {
		if err := lumberjackWriter.Close(); err != nil {
			return err
		}
		lumberjackWriter = nil
	}

	// 关闭 rotateLogs writer（如果存在）
	if rotateLogsWriter != nil {
		// rotatelogs.RotateLogs 有 Close() 方法
		if err := rotateLogsWriter.Close(); err != nil {
			return err
		}
		rotateLogsWriter = nil
	}

	// 清理 writer 引用
	fileWriter = nil

	// 关闭 WQGroup/logger
	return wqlogger.Close()
}
