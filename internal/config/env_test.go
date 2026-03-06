package config

import (
	"strings"
	"testing"
)

func Test_enrichReplicasFromEnv(t *testing.T) {
	t.Setenv("REPLICA0_URL", "https://origin-env:443")
	_, err := enrichReplicasFromEnv(nil)

	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "numbered replica env variables must have a number id >= 1") {
		t.Errorf("expected error to contain 'numbered replica env variables must have a number id >= 1' but got '%v'", err)
	}
}
