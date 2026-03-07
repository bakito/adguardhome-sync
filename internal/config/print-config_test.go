package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bakito/adguardhome-sync/internal/types"
	"github.com/bakito/adguardhome-sync/version"
)

func TestAppConfig_PrintInternal(t *testing.T) {
	env := []string{"FOO=foo", "BAR=bar"}

	t.Run("without file", func(t *testing.T) {
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
		}
		out, err := ac.printInternal(env, "v0.0.1", []string{"v0.0.2"})
		if err != nil {
			t.Fatalf("printInternal error = %v, want nil", err)
		}
		expectedStr := fmt.Sprintf(expected(t, 1), version.Version, version.Build, runtime.GOOS, runtime.GOARCH)
		if diff := diffIgnoringLineEndings(expectedStr, out); diff != "" {
			t.Errorf("printInternal mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("with file", func(t *testing.T) {
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
			filePath: "config.yaml",
		}
		out, err := ac.printInternal(env, "v0.0.1", []string{"v0.0.2"})
		if err != nil {
			t.Fatalf("printInternal error = %v, want nil", err)
		}
		expectedStr := fmt.Sprintf(expected(t, 2), version.Version, version.Build, runtime.GOOS, runtime.GOARCH)
		if diff := diffIgnoringLineEndings(expectedStr, out); diff != "" {
			t.Errorf("printInternal mismatch (-want +got):\n%s", diff)
		}
	})
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

func diffIgnoringLineEndings(expected, actual string) string {
	normalizedActual := strings.ReplaceAll(actual, "\r\n", "\n")
	normalizedExpected := strings.ReplaceAll(expected, "\r\n", "\n")
	return cmp.Diff(normalizedExpected, normalizedActual)
}
