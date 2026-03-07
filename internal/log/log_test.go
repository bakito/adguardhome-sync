package log

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLogListClone(t *testing.T) {
	tests := []struct {
		name string
		enc  zapcore.Encoder
	}{
		{"valid console encoder", zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())},
		{"valid JSON encoder", zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ll := &logList{
				LevelEnabler: zapcore.DebugLevel,
				enc:          tt.enc,
			}
			cloned := ll.clone()
			if cloned.LevelEnabler != ll.LevelEnabler {
				t.Error("LevelEnabler mismatch between original and clone")
			}
			if cloned.enc == nil {
				t.Error("Encoder mismatch between original and clone")
			}
		})
	}
}

func TestLogListWith(t *testing.T) {
	tests := []struct {
		name   string
		fields []zapcore.Field
	}{
		{"no fields", nil},
		{"with fields", []zapcore.Field{zap.String("key", "value")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ll := &logList{
				LevelEnabler: zapcore.DebugLevel,
				enc:          zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
			}
			core := ll.With(tt.fields)
			if core == nil {
				t.Fatal("With() returned nil core")
			}
		})
	}
}

func TestLogListCheck(t *testing.T) {
	tests := []struct {
		name    string
		level   zapcore.Level
		enabled bool
	}{
		{"enabled debug level", zapcore.DebugLevel, true},
		{"disabled error level", zapcore.ErrorLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ll := &logList{
				LevelEnabler: zap.LevelEnablerFunc(func(l zapcore.Level) bool {
					return l == tt.level && tt.enabled
				}),
				enc: zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
			}
			entry := zapcore.Entry{Level: tt.level}
			ce := ll.Check(entry, nil)
			if (ce != nil) != tt.enabled {
				t.Errorf("Check() returned %v, want %v", ce != nil, tt.enabled)
			}
		})
	}
}

func TestLogListWrite(t *testing.T) {
	tests := []struct {
		name    string
		fields  []zapcore.Field
		wantErr bool
	}{
		{"no fields", nil, false},
		{"with field", []zapcore.Field{zap.String("key", "value")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ll := &logList{
				LevelEnabler: zapcore.DebugLevel,
				enc:          zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
			}
			logs = nil
			err := ll.Write(zapcore.Entry{Message: "test"}, tt.fields)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(logs) == 0 || logs[0] == "" {
				t.Error("Write() did not append logs")
			}
		})
	}
}

func TestLogListSync(t *testing.T) {
	ll := &logList{}
	if err := ll.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger("test")
	if logger == nil {
		t.Fatal("GetLogger() returned nil")
	}
	if logger.Desugar() == nil {
		t.Fatal("GetLogger().Desugar() returned nil")
	}
}

func TestLogs(t *testing.T) {
	logs = []string{"log1", "log2"}
	retrieved := Logs()
	if len(retrieved) != 2 {
		t.Errorf("Logs() = %v, want %v", len(retrieved), 2)
	}
}

func TestClear(t *testing.T) {
	logs = []string{"log1", "log2"}
	Clear()
	if len(logs) != 0 {
		t.Errorf("Clear() left logs with length %d", len(logs))
	}
}

func TestInit(t *testing.T) {
	t.Setenv(envLogLevel, "debug")
	t.Setenv(envLogFormat, "json")

	logger, err := initRootLogger()
	if err != nil {
		t.Fatalf("initRootLogger() error = %v", err)
	}
	if logger == nil {
		t.Fatal("initRootLogger() returned nil")
	}
}
