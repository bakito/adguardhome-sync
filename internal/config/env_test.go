package config

import (
	"strings"
	"testing"
)

func TestEnrichReplicasFromEnv(t *testing.T) {
	t.Setenv("REPLICA0_URL", "https://origin-env:443")

	_, err := enrichReplicasFromEnv(nil)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "numbered replica env variables must have a number id >= 1") {
		t.Errorf("expected error containing 'numbered replica env variables must have a number id >= 1', got '%s'", err.Error())
	}
}
