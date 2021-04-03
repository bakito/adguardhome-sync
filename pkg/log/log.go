package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const logHistorySize = 50

var (
	rootLogger *zap.Logger
	logs       []string
)

// GetLogger returns a named logger
func GetLogger(name string) *zap.SugaredLogger {
	return rootLogger.Named(name).Sugar()
}

func init() {
	level := zap.InfoLevel

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	opt := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, &logList{enc: zapcore.NewConsoleEncoder(cfg.EncoderConfig)})
	})

	rootLogger, _ = cfg.Build(opt)
}

type logList struct {
	enc zapcore.Encoder
}

func (l *logList) Enabled(_ zapcore.Level) bool {
	return true
}

func (l *logList) With(_ []zapcore.Field) zapcore.Core {
	return l
}

func (l *logList) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, l)
}

func (l *logList) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := l.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	logs = append(logs, buf.String())

	if len(logs) > logHistorySize {
		logs = logs[len(logs)-logHistorySize:]
	}
	return nil
}

func (l *logList) Sync() error {
	return nil
}

func Logs() []string {
	var list []string
	for _, l := range logs {
		list = append(list, l)
	}
	return list
}
