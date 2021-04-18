package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logHistorySize = 50
	envLogLevel    = "LOG_LEVEL"
)

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

	if lvl, ok := os.LookupEnv(envLogLevel); ok {
		if err := level.Set(lvl); err != nil {
			panic(err)
		}
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	opt := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, &logList{
			enc:          zapcore.NewConsoleEncoder(cfg.EncoderConfig),
			LevelEnabler: cfg.Level,
		})
	})

	rootLogger, _ = cfg.Build(opt)
}

type logList struct {
	zapcore.LevelEnabler
	enc zapcore.Encoder
}

func (l *logList) clone() *logList {
	return &logList{
		LevelEnabler: l.LevelEnabler,
		enc:          l.enc.Clone(),
	}
}

func (l *logList) With(fields []zapcore.Field) zapcore.Core {
	clone := l.clone()
	addFields(clone.enc, fields)
	return clone
}

func (l *logList) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if l.Enabled(ent.Level) {
		return ce.AddCore(ent, l)
	}
	return ce
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

// Logs get the current logs
func Logs() []string {
	return logs
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}
