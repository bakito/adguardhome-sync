package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bakito/adguardhome-sync/internal/types"
	"github.com/bakito/adguardhome-sync/version"

	_ "embed"
)

func TestAppConfig_printInternal(t *testing.T) {
	env := []string{"FOO=foo", "BAR=bar"}

	tests := []struct {
		name     string
		filePath string
		expected int
	}{
		{
			name:     "should printInternal config without file",
			filePath: "",
			expected: 1,
		},
		{
			name:     "should printInternal config with file",
			filePath: "config.yaml",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &AppConfig{
				cfg: &types.Config{
					Origin: &types.AdGuardInstance{
						URL: "https://ha.xxxx.net:3000",
					},
				},
				content: `
origin:
  url: https://ha.xxxx.net:3000
`,
				filePath: tt.filePath,
			}

			out, err := ac.printInternal(env, "v0.0.1", []string{"v0.0.2"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			want := fmt.Sprintf(expected(t, tt.expected), version.Version, version.Build, runtime.GOOS, runtime.GOARCH)
			if normalize(out) != normalize(want) {
				t.Errorf("expected %s but got %s", want, out)
			}
		})
	}
}

func normalize(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

func expected(t *testing.T, id int) string {
	t.Helper()
	b, err := os.ReadFile(
		filepath.Join("..", "..", "testdata", "config", fmt.Sprintf("print-config_test_expected%d.md", id)),
	)
	if err != nil {
		t.Fatalf("failed to read expected file: %v", err)
	}
	return string(b)
}
