package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func Test_HealthcheckCommand_Flags(t *testing.T) {
	if healthcheckCmd == nil {
		t.Fatal("healthcheckCmd should not be nil")
	}
	if healthcheckCmd.Use != "healthcheck" {
		t.Errorf("healthcheckCmd.Use = %v, want healthcheck", healthcheckCmd.Use)
	}

	foundAlias := false
	for _, alias := range healthcheckCmd.Aliases {
		if alias == "health" {
			foundAlias = true
			break
		}
	}
	if !foundAlias {
		t.Errorf("healthcheckCmd.Aliases = %v, want to contain 'health'", healthcheckCmd.Aliases)
	}

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{"port", "p", "0"},
		{"url", "u", ""},
		{"timeout", "t", "5s"},
		{"insecure-skip-verify", "k", "false"},
		{"ready", "r", "false"},
		{"api-port", "", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := healthcheckCmd.Flags().Lookup(tt.name)
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

func Test_HealthcheckCommand_Execution(t *testing.T) {
	t.Run("should succeed when server returns 200 OK via custom URL", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/healthz" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		cmd := &cobra.Command{}
		cmd.SetContext(context.Background())
		cmd.Flags().StringVarP(&healthcheckURL, "url", "u", ts.URL+"/healthz", "")
		cmd.Flags().DurationVarP(&healthcheckTimeout, "timeout", "t", 2*time.Second, "")
		cmd.Flags().BoolVarP(&healthcheckInsecureSkipVerify, "insecure-skip-verify", "k", false, "")

		err := healthcheckCmd.RunE(cmd, nil)
		if err != nil {
			t.Errorf("RunE() error = %v, want nil", err)
		}
	})

	t.Run("should fail when server returns 503 Service Unavailable", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer ts.Close()

		cmd := &cobra.Command{}
		cmd.SetContext(context.Background())
		cmd.Flags().StringVarP(&healthcheckURL, "url", "u", ts.URL+"/healthz", "")
		cmd.Flags().DurationVarP(&healthcheckTimeout, "timeout", "t", 2*time.Second, "")
		cmd.Flags().BoolVarP(&healthcheckInsecureSkipVerify, "insecure-skip-verify", "k", false, "")

		err := healthcheckCmd.RunE(cmd, nil)
		if err == nil {
			t.Error("RunE() error = nil, want error")
		}
	})

	t.Run("should succeed when targeting custom port (defaults to livez)", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/livez" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		parsedURL, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		port, err := strconv.Atoi(parsedURL.Port())
		if err != nil {
			t.Fatal(err)
		}

		healthcheckURL = ""
		healthcheckPort = port
		healthcheckTimeout = 2 * time.Second
		healthcheckInsecureSkipVerify = false
		healthcheckReady = false

		cmd := &cobra.Command{}
		cmd.SetContext(context.Background())
		cmd.Flags().IntVarP(&healthcheckPort, "port", "p", port, "")
		cmd.Flags().StringVarP(&healthcheckURL, "url", "u", "", "")
		cmd.Flags().DurationVarP(&healthcheckTimeout, "timeout", "t", 2*time.Second, "")
		cmd.Flags().BoolVarP(&healthcheckInsecureSkipVerify, "insecure-skip-verify", "k", false, "")
		cmd.Flags().BoolVarP(&healthcheckReady, "ready", "r", false, "")

		err = healthcheckCmd.RunE(cmd, nil)
		if err != nil {
			t.Errorf("RunE() error = %v, want nil", err)
		}
	})

	t.Run("should succeed when targeting custom port with ready flag (targets readyz)", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/readyz" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		parsedURL, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		port, err := strconv.Atoi(parsedURL.Port())
		if err != nil {
			t.Fatal(err)
		}

		healthcheckURL = ""
		healthcheckPort = port
		healthcheckTimeout = 2 * time.Second
		healthcheckInsecureSkipVerify = false
		healthcheckReady = true

		cmd := &cobra.Command{}
		cmd.SetContext(context.Background())
		cmd.Flags().IntVarP(&healthcheckPort, "port", "p", port, "")
		cmd.Flags().StringVarP(&healthcheckURL, "url", "u", "", "")
		cmd.Flags().DurationVarP(&healthcheckTimeout, "timeout", "t", 2*time.Second, "")
		cmd.Flags().BoolVarP(&healthcheckInsecureSkipVerify, "insecure-skip-verify", "k", false, "")
		cmd.Flags().BoolVarP(&healthcheckReady, "ready", "r", true, "")

		err = healthcheckCmd.RunE(cmd, nil)
		if err != nil {
			t.Errorf("RunE() error = %v, want nil", err)
		}
	})

	t.Run("should fail when API port is 0", func(t *testing.T) {
		healthcheckURL = ""
		healthcheckPort = 0
		healthcheckTimeout = 2 * time.Second

		cmd := &cobra.Command{}
		cmd.SetContext(context.Background())
		cmd.Flags().Int("api-port", 0, "")
		_ = cmd.Flags().Set("api-port", "0")

		err := healthcheckCmd.RunE(cmd, nil)
		if err == nil {
			t.Error("RunE() error = nil, want error for port 0")
		}
	})

	t.Run("should fail when server is not reachable", func(t *testing.T) {
		healthcheckURL = "http://127.0.0.1:64999/healthz"
		healthcheckTimeout = 100 * time.Millisecond

		cmd := &cobra.Command{}
		cmd.SetContext(context.Background())
		cmd.Flags().StringVarP(&healthcheckURL, "url", "u", healthcheckURL, "")
		cmd.Flags().DurationVarP(&healthcheckTimeout, "timeout", "t", healthcheckTimeout, "")

		err := healthcheckCmd.RunE(cmd, nil)
		if err == nil {
			t.Error("RunE() error = nil, want connection error")
		}
	})
}
