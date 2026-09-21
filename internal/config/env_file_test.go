package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSecret(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// unsetOnCleanup removes vars that resolveEnvFiles sets via os.Setenv, as t.Setenv does not track them.
func unsetOnCleanup(t *testing.T, names ...string) {
	t.Helper()
	t.Cleanup(func() {
		for _, n := range names {
			_ = os.Unsetenv(n)
		}
	})
}

func TestResolveEnvFiles(t *testing.T) {
	unsetOnCleanup(t, "ORIGIN_PASSWORD", "REPLICA_COOKIE", "REPLICA2_USERNAME", "REPLICA2_FEATURES_THEME",
		"API_PASSWORD", "CRON")
	t.Setenv("ORIGIN_PASSWORD_FILE", writeSecret(t, "origin-pw\n"))
	t.Setenv("REPLICA_COOKIE_FILE", writeSecret(t, "name=value\r\n"))
	t.Setenv("REPLICA2_USERNAME_FILE", writeSecret(t, "user2"))
	t.Setenv("REPLICA2_FEATURES_THEME_FILE", writeSecret(t, "false"))
	t.Setenv("API_PASSWORD_FILE", writeSecret(t, "api-pw"))
	t.Setenv("CRON_FILE", writeSecret(t, "* * * * *"))
	// unknown names must be ignored, even if the file does not exist
	t.Setenv("SSL_CERT_FILE", "/does/not/exist")

	if err := resolveEnvFiles(); err != nil {
		t.Fatal(err)
	}

	for k, want := range map[string]string{
		"ORIGIN_PASSWORD":         "origin-pw",
		"REPLICA_COOKIE":          "name=value",
		"REPLICA2_USERNAME":       "user2",
		"REPLICA2_FEATURES_THEME": "false",
		"API_PASSWORD":            "api-pw",
		"CRON":                    "* * * * *",
	} {
		if got := os.Getenv(k); got != want {
			t.Errorf("%s: want %q, got %q", k, want, got)
		}
	}
	if _, ok := os.LookupEnv("SSL_CERT"); ok {
		t.Error("SSL_CERT must not be set")
	}
}

func TestResolveEnvFiles_Conflict(t *testing.T) {
	t.Setenv("ORIGIN_PASSWORD", "a")
	t.Setenv("ORIGIN_PASSWORD_FILE", writeSecret(t, "b"))
	if err := resolveEnvFiles(); err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveEnvFiles_Missing(t *testing.T) {
	t.Setenv("ORIGIN_PASSWORD_FILE", "/does/not/exist")
	if err := resolveEnvFiles(); err == nil {
		t.Fatal("expected error")
	}
}

func TestGet_EnvFile(t *testing.T) {
	unsetOnCleanup(t, "ORIGIN_USERNAME", "ORIGIN_PASSWORD", "REPLICA1_COOKIE")
	t.Setenv("ORIGIN_URL", "https://origin:443")
	t.Setenv("ORIGIN_USERNAME_FILE", writeSecret(t, "ou\n"))
	t.Setenv("ORIGIN_PASSWORD_FILE", writeSecret(t, "op\n"))
	t.Setenv("REPLICA1_URL", "https://replica:443")
	t.Setenv("REPLICA1_COOKIE_FILE", writeSecret(t, "c=v\n"))

	cfg, err := Get("", nil)
	if err != nil {
		t.Fatal(err)
	}
	c := cfg.Get()
	if c.Origin.Username != "ou" || c.Origin.Password != "op" {
		t.Errorf("origin: got %q/%q", c.Origin.Username, c.Origin.Password)
	}
	if len(c.Replicas) != 1 || c.Replicas[0].Cookie != "c=v" {
		t.Errorf("replicas: got %+v", c.Replicas)
	}
}
