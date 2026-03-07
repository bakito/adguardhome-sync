package cmd

import (
	"testing"
)

func Test_RootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
	if rootCmd.Use != "adguardhome-sync" {
		t.Errorf("rootCmd.Use = %v, want adguardhome-sync", rootCmd.Use)
	}

	// Check persistent flags
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Fatal("config flag not found")
	}
	expectedUsage := "config file (default is $HOME/.adguardhome-sync.yaml)"
	if configFlag.Usage != expectedUsage {
		t.Errorf("configFlag.Usage = %v, want %v", configFlag.Usage, expectedUsage)
	}
}

func Test_RunCommand(t *testing.T) {
	if doCmd == nil {
		t.Fatal("doCmd should not be nil")
	}
	if doCmd.Use != "run" {
		t.Errorf("doCmd.Use = %v, want run", doCmd.Use)
	}

	// Check some representative flags
	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{"cron", "", ""},
		{"runOnStart", "", "true"},
		{"printConfigOnly", "", "false"},
		{"api-port", "", "8080"},
		{"origin-url", "", ""},
		{"replica-url", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := doCmd.PersistentFlags().Lookup(tt.name)
			if flag == nil {
				// Try non-persistent just in case
				flag = doCmd.Flags().Lookup(tt.name)
			}
			if flag == nil {
				t.Fatalf("Flag %s not found", tt.name)
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s shorthand = %v, want %v", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("Flag %s default value = %v, want %v", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}
