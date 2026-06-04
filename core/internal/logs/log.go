package logs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Logger struct {
	*zap.Logger
	level     zapcore.Level
	FirstName string
	option    Option
}

type Option struct {
	Level          string `yaml:"Level"`          //日志等级
	FilePath       string `yaml:"FilePath"`       //文件目录
	ShowConsoleLog bool   `yaml:"ShowConsoleLog"` //是否日志展出
	WriteFile      bool   `yaml:"WriteFile"`      //是否写出到文件
	MoreFiles      bool   `yaml:"MoreFiles"`      //是否将各种类型分开写入文件
	FirstName      string `yaml:"FirstName"`      //日志前置
}

var DefaultOption = Option{
	Level:          "debug",
	FilePath:       "log",
	WriteFile:      true,
	MoreFiles:      true,
	ShowConsoleLog: true,
	FirstName:      "main",
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	//enc.AppendString(t.Format("2006-01-02 15:04:05"))
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
func NewLog(cfg Option) *Logger {
	// 解析日志级别
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}
	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	// 构建输出目标
	var cores []zapcore.Core

	// 控制台输出
	if cfg.ShowConsoleLog {
		consoleWriter := zapcore.AddSync(os.Stdout)
		cores = append(cores, zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			consoleWriter,
			level,
		))
	}
	// 文件输出
	if cfg.WriteFile {
		if cfg.MoreFiles {
			// 按级别分开写入不同文件
			levelFiles := map[zapcore.Level]string{
				zapcore.DebugLevel:  fmt.Sprintf("%s-%s.log", cfg.FirstName, "debug"),
				zapcore.InfoLevel:   fmt.Sprintf("%s-%s.log", cfg.FirstName, "info"),
				zapcore.WarnLevel:   fmt.Sprintf("%s-%s.log", cfg.FirstName, "warn"),
				zapcore.ErrorLevel:  fmt.Sprintf("%s-%s.log", cfg.FirstName, "error"),
				zapcore.DPanicLevel: fmt.Sprintf("%s-%s.log", cfg.FirstName, "error"),
				zapcore.PanicLevel:  fmt.Sprintf("%s-%s.log", cfg.FirstName, "error"),
				zapcore.FatalLevel:  fmt.Sprintf("%s-%s.log", cfg.FirstName, "error"),
			}
			for lv, filename := range levelFiles {
				writer, err := getLogWriter(cfg.FilePath+"/"+time.Now().Format(time.DateOnly), filename)
				if err != nil {
					continue
				}
				cores = append(cores, zapcore.NewCore(
					zapcore.NewJSONEncoder(encoderConfig),
					writer,
					zap.LevelEnablerFunc(func(l zapcore.Level) bool {
						return l == lv
					}),
				))
			}
		} else {
			// 全部写入同一个文件
			writer, err := getLogWriter(cfg.FilePath+"/"+time.Now().Format(time.DateOnly), fmt.Sprintf("%s.log", cfg.FirstName))
			if err == nil {
				cores = append(cores, zapcore.NewCore(
					zapcore.NewJSONEncoder(encoderConfig),
					writer,
					level,
				))
			}
		}
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &Logger{Logger: logger, level: level, FirstName: cfg.FirstName, option: cfg}
}

func getLogWriter(filePath, filename string) (zapcore.WriteSyncer, error) {
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}
	file, err := os.OpenFile(
		filepath.Join(filePath, filename),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	return zapcore.AddSync(file), nil
}

func (l *Logger) Log(level zapcore.Level, format string, data []map[string]interface{}) {
	var ls []zap.Field
	for _, vv := range data {
		for k, v := range vv {
			switch d := v.(type) {
			case string:
				ls = append(ls, zap.String(k, d))
			case error:
				ls = append(ls, zap.Error(d))
			default:
				ls = append(ls, zap.Any(k, v))
			}
		}
	}

	l.Logger.Log(level, format, ls...)
}

func (l *Logger) Info(msg string, data ...map[string]interface{}) {
	l.Log(zapcore.InfoLevel, msg, data)
}

func (l *Logger) Warn(msg string, data ...map[string]interface{}) {
	l.Log(zapcore.WarnLevel, msg, data)
}

func (l *Logger) Error(msg string, data ...map[string]interface{}) {
	l.Log(zapcore.ErrorLevel, msg, data)
}

func (l *Logger) Debug(msg string, data ...map[string]interface{}) {
	l.Log(zapcore.DebugLevel, msg, data)
}

func (l *Logger) Fatal(msg string, data ...map[string]interface{}) {
	l.Log(zapcore.FatalLevel, msg, data)
}

func (l *Logger) Panic(msg string, data ...map[string]interface{}) {
	l.Log(zapcore.PanicLevel, msg, data)
}

type GormLogger struct {
	*Logger
	// 慢查询阈值
	slowThreshold time.Duration
	// 是否启用 SQL 日志
	enable bool
}

// NewGormLogger 创建 GORM 日志适配器
func (l *Logger) NewGormLogger(slowThreshold time.Duration, enable bool) *GormLogger {
	return &GormLogger{
		Logger:        l,
		slowThreshold: slowThreshold,
		enable:        enable,
	}
}

// LogMode 实现 gorm logger 接口
func (g *GormLogger) LogMode(logLevel gormlogger.LogLevel) gormlogger.Interface {
	newLogger := g.Logger.Clone()
	switch logLevel {
	case gormlogger.Silent:
		newLogger.level = zapcore.FatalLevel + 1
	case gormlogger.Error:
		newLogger.level = zapcore.ErrorLevel
	case gormlogger.Warn:
		newLogger.level = zapcore.WarnLevel
	case gormlogger.Info:
		newLogger.level = zapcore.InfoLevel
	}
	return &GormLogger{
		Logger:        newLogger,
		slowThreshold: g.slowThreshold,
		enable:        g.enable,
	}
}

// Info 实现 gorm logger 接口
func (g *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	g.Logger.Sugar().Infof(msg, data...)
}

// Warn 实现 gorm logger 接口
func (g *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	g.Logger.Sugar().Warnf(msg, data...)
}

// Error 实现 gorm logger 接口
func (g *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	g.Logger.Sugar().Errorf(msg, data...)
}

// Trace 实现 gorm logger 接口
func (g *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !g.enable {
		return
	}

	latency := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		g.Logger.Logger.Error(sql, zap.String("tag", "sql"), zap.Error(err), zap.Duration("latency", latency))
	case g.slowThreshold != 0 && latency > g.slowThreshold:
		g.Logger.Logger.Warn(sql, zap.String("tag", "sql"), zap.Duration("latency", latency), zap.Int64("rows", rows))
	default:
		if g.level <= zapcore.DebugLevel {
			g.Logger.Logger.Debug(sql, zap.String("tag", "sql"), zap.Duration("latency", latency), zap.Int64("rows", rows))
		}
	}
}

// Clone 克隆 Logger，支持传入新的 FirstName 来重建文件写入器
func (l *Logger) Clone(firstName ...string) *Logger {
	fn := l.FirstName
	if len(firstName) > 0 {
		fn = firstName[0]
	}

	// 如果 FirstName 没变，直接浅克隆
	if fn == l.FirstName {
		return &Logger{
			Logger:    l.Logger.WithOptions(),
			level:     l.level,
			FirstName: fn,
			option:    l.option,
		}
	}

	// FirstName 变了，需要重建文件写入器
	newOpt := l.option
	newOpt.FirstName = fn
	return NewLog(newOpt)
}

// WithFirstName 返回带新 FirstName 的新 Logger 实例（会重建文件写入器）
func (l *Logger) WithFirstName(name string) *Logger {
	return l.Clone(name)
}
