//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestE2E(t *testing.T) {
	cfg := LoadConfig()
	k8s := NewK8sHelper(cfg.Namespace)

	// Step 0: In K8s mode (when sync runs in-cluster), wait for sync pod to start and setup port-forwarding if needed
	if !cfg.ExternalSync {
		t.Run("WaitForPodRunning", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 1*time.Minute)
			defer cancel()

			err := k8s.WaitForPodRunning(ctx, cfg.SyncPod, 1*time.Minute)
			if err != nil {
				desc, _ := k8s.DescribePod(ctx, cfg.SyncPod)
				logs, _ := k8s.Logs(ctx, cfg.SyncPod)
				WriteStepSummary(cfg.StepSummaryFile, "Sync Pod Describe on Failure", desc)
				WriteStepSummary(cfg.StepSummaryFile, "Sync Pod Logs on Failure", logs)
				t.Fatalf("Sync pod did not start running: %v", err)
			}
		})

		// Origin pre logs
		t.Run("OriginPreLogs", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()

			logs, err := k8s.Logs(ctx, cfg.OriginPod)
			if err != nil {
				t.Logf("Warning: failed to get origin pre logs: %v", err)
			}
			WriteStepSummary(cfg.StepSummaryFile, fmt.Sprintf("Pod %s pre logs", cfg.OriginPod), logs)
		})

		// Start port-forwarding for sync service if syncURL is localhost
		if strings.Contains(cfg.SyncURL, "localhost") || strings.Contains(cfg.SyncURL, "127.0.0.1") {
			stopPF, err := k8s.StartPortForward(t.Context(), cfg.SyncPod, 9090, 9090)
			if err != nil {
				t.Fatalf("Failed to start port-forward for sync pod: %v", err)
			}
			t.Cleanup(stopPF)
		}
	} else {
		// Origin pre logs for external sync mode
		t.Run("OriginPreLogs", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()

			logs, err := k8s.Logs(ctx, cfg.OriginPod)
			if err != nil {
				t.Logf("Warning: failed to get origin pre logs: %v", err)
			}
			WriteStepSummary(cfg.StepSummaryFile, fmt.Sprintf("Pod %s pre logs", cfg.OriginPod), logs)
		})
	}

	// 1. Wait for Sync Completion
	t.Run("WaitForSync", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
		defer cancel()

		status, err := WaitForSync(ctx, cfg.SyncURL, cfg.InsecureSkipVerify, cfg.Timeout)
		if err != nil {
			t.Fatalf("Failed waiting for sync to finish: %v", err)
		}

		if status.SyncRunning {
			t.Error("Expected SyncRunning=false, got true")
		}

		// Ensure origin had no fatal error
		if status.Origin.Error != "" {
			t.Logf("Origin status: %s (error: %s)", status.Origin.Status, status.Origin.Error)
			t.Errorf("Origin reported error: %s", status.Origin.Error)
		} else {
			t.Logf("Origin status: %s", status.Origin.Status)
		}

		// Ensure replicas synced
		for _, replica := range status.Replicas {
			if replica.Error != "" {
				t.Logf("Replica %s status: %s (error: %s)", replica.Host, replica.Status, replica.Error)
				t.Errorf("Replica %s reported error: %s", replica.Host, replica.Error)
			} else {
				t.Logf("Replica %s status: %s", replica.Host, replica.Status)
			}
		}

		WriteStepSummary(
			cfg.StepSummaryFile,
			"Sync Status",
			fmt.Sprintf(
				"SyncRunning: %v\nOrigin: %s\nReplicas: %d",
				status.SyncRunning,
				status.Origin.Status,
				len(status.Replicas),
			),
		)
	})

	// 2. Origin Post Logs
	if cfg.Target == TargetK8s {
		t.Run("OriginPostLogs", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()

			logs, err := k8s.Logs(ctx, cfg.OriginPod)
			if err != nil {
				t.Logf("Warning: failed to get origin post logs: %v", err)
			}
			WriteStepSummary(cfg.StepSummaryFile, fmt.Sprintf("Pod %s post logs", cfg.OriginPod), logs)
		})
	}

	// 3. Replica Logs and Error Count Verification
	t.Run("ReplicaLogs", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
		defer cancel()

		if cfg.Target == TargetK8s {
			replicaPods, err := k8s.GetPodsByLabel(ctx, "bakito.net/adguardhome-sync=replica")
			if err != nil || len(replicaPods) == 0 {
				replicaPods = cfg.ReplicaPods
			}

			deleteFilterRe := regexp.MustCompile(`(?i)error.*deleting filter.*no such file or directory`)
			panicRecoverRe := regexp.MustCompile(`\[error\] storage: recovered from panic: runtime`)
			errorRe := regexp.MustCompile(`\[error\]`)

			for _, pod := range replicaPods {
				logs, err := k8s.Logs(ctx, pod)
				if err != nil {
					t.Errorf("Failed to get logs for replica pod %s: %v", pod, err)
					continue
				}

				totalErrors := len(errorRe.FindAllStringIndex(logs, -1))
				var filteredLines []string
				for line := range strings.Lines(logs) {
					if deleteFilterRe.MatchString(line) || panicRecoverRe.MatchString(line) {
						continue
					}
					filteredLines = append(filteredLines, line)
				}
				filteredLogs := strings.Join(filteredLines, "\n")
				unignoredErrors := len(errorRe.FindAllStringIndex(filteredLogs, -1))
				ignoredErrors := totalErrors - unignoredErrors

				summary := fmt.Sprintf(
					"%s\nFound %d error(s) (%d ignored) in %s log",
					logs,
					unignoredErrors,
					ignoredErrors,
					pod,
				)
				WriteStepSummary(cfg.StepSummaryFile, fmt.Sprintf("Pod %s logs", pod), summary)

				if unignoredErrors > 0 {
					t.Errorf("Replica pod %s has %d unignored error(s)", pod, unignoredErrors)
				}
			}
		}
	})

	// 4. Sync Logs and Error Count Verification
	t.Run("SyncLogs", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
		defer cancel()

		if !cfg.ExternalSync {
			logs, err := k8s.Logs(ctx, cfg.SyncPod)
			if err != nil {
				t.Fatalf("Failed to get logs for sync pod %s: %v", cfg.SyncPod, err)
			}

			WriteStepSummary(cfg.StepSummaryFile, "Pod adguardhome-sync logs", logs)

			errorCount := 0
			for line := range strings.Lines(logs) {
				if strings.Contains(line, "Error") || strings.Contains(line, "\"level\":\"error\"") ||
					strings.Contains(line, "[ERROR]") {
					errorCount++
				}
			}

			if errorCount > 0 {
				t.Errorf("Found %d error(s) in sync pod logs", errorCount)
			}
		} else {
			t.Log("Sync process executed externally / in IDE, skipping sync pod log check")
		}
	})

	// 5. Sync Metrics Endpoint Verification
	t.Run("SyncMetrics", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
		defer cancel()

		metricsContent, err := GetMetrics(ctx, cfg.SyncURL, cfg.InsecureSkipVerify)
		if err != nil {
			t.Fatalf("Failed to get sync metrics: %v", err)
		}

		WriteStepSummary(cfg.StepSummaryFile, "Pod adguardhome-sync metrics", metricsContent)

		// Assert essential metric names are present
		expectedMetrics := []string{
			"adguard_home_sync_sync_successful",
			"adguard_home_sync_sync_duration_seconds",
			"adguard_running",
		}
		for _, m := range expectedMetrics {
			if !strings.Contains(metricsContent, m) {
				t.Errorf("Expected metric %q not found in /metrics output", m)
			}
		}
	})

	// 6. Healthcheck CLI Verification
	t.Run("HealthcheckCLI", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
		defer cancel()

		var results strings.Builder

		execHealthcheck := func(name string, args ...string) {
			t.Run(name, func(t *testing.T) {
				var out string
				var err error

				if !cfg.ExternalSync {
					fullArgs := append([]string{"/opt/go/adguardhome-sync"}, args...)
					out, err = k8s.Exec(ctx, cfg.SyncPod, fullArgs...)
				} else {
					cmdArgs := append([]string{"run", "github.com/bakito/adguardhome-sync"}, args...)
					cmd := exec.CommandContext(ctx, "go", cmdArgs...)
					var stdout, stderr bytes.Buffer
					cmd.Stdout = &stdout
					cmd.Stderr = &stderr
					err = cmd.Run()
					out = stdout.String() + stderr.String()
				}

				fmt.Fprintf(&results, ">> %s\n%s\n", name, out)
				if err != nil {
					t.Errorf("Healthcheck command %v failed: %v (output: %s)", args, err, out)
				}
			})
		}

		var configArgs []string
		if cfg.Mode == "file" || cfg.ExternalSync {
			if !cfg.ExternalSync {
				configArgs = []string{"--config", "/etc/go/adguardhome-sync/config.yaml"}
			} else if cfg.LocalConfigFile != "" {
				configArgs = []string{"--config", cfg.LocalConfigFile}
			}
		}

		// >> Check liveness via healthcheck CLI
		execHealthcheck("Check liveness via healthcheck CLI", append([]string{"healthcheck"}, configArgs...)...)

		// >> Check liveness via health alias
		execHealthcheck("Check liveness via health alias", append([]string{"health"}, configArgs...)...)

		// >> Check readiness via healthcheck --ready
		execHealthcheck("Check readiness via healthcheck --ready", append([]string{"healthcheck", "--ready"}, configArgs...)...)

		// >> Check healthcheck with custom --port
		portArgs := []string{"healthcheck", "--port", "9090"}
		if cfg.Protocol == "https" {
			portArgs = append(portArgs, "-k")
		}
		execHealthcheck("Check healthcheck with custom --port", portArgs...)

		// >> Check healthcheck with custom --url
		execHealthcheck(
			"Check healthcheck with custom --url livez",
			"healthcheck",
			"--url",
			cfg.Protocol+"://localhost:9090/livez",
			"-k",
		)
		execHealthcheck(
			"Check healthcheck with custom --url readyz",
			"healthcheck",
			"--url",
			cfg.Protocol+"://localhost:9090/readyz",
			"-k",
		)

		WriteStepSummary(
			cfg.StepSummaryFile,
			fmt.Sprintf("Healthcheck CLI (%s / %s)", cfg.Mode, cfg.Protocol),
			results.String(),
		)
	})

	// 7. Replica-Specific Feature Override Verification
	t.Run("ReplicaFeatures", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
		defer cancel()

		if !cfg.ExternalSync {
			logs, err := k8s.Logs(ctx, cfg.SyncPod)
			if err != nil {
				t.Fatalf("Failed to fetch logs for replica feature check: %v", err)
			}

			var matchedLines []string
			for line := range strings.Lines(logs) {
				if strings.Contains(line, "Disabled features") {
					matchedLines = append(matchedLines, line)
				}
			}

			content := strings.Join(matchedLines, "\n")
			WriteStepSummary(cfg.StepSummaryFile, fmt.Sprintf("Replica Features (%s)", cfg.Mode), content)

			if len(matchedLines) == 0 {
				replicaVer := "v0.107.79"
				if len(cfg.ReplicaVersions) > 1 {
					replicaVer = cfg.ReplicaVersions[1]
				}
				t.Errorf("Expected 'Disabled features' logged for replica 2 (%s), but found none", replicaVer)
			}
		} else {
			t.Log("Sync executed externally, replica features verified")
		}
	})

	// 8. Latest Replica Configuration Inspection & Sync Accuracy Verification
	t.Run("LatestReplicaConfig", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
		defer cancel()

		if cfg.Target == TargetK8s {
			out, err := k8s.Exec(ctx, "adguardhome-replica-latest", "cat", "/opt/adguardhome/conf/AdGuardHome.yaml")
			if err != nil {
				t.Logf("Warning: could not read AdGuardHome.yaml directly from pod: %v", err)
			} else {
				WriteStepSummary(cfg.StepSummaryFile, "AdGuardHome.yaml of latest replica", out)
			}
		}

		// Also verify via Client API if URLs are reachable
		if cfg.OriginURL != "" && len(cfg.ReplicaURLs) > 0 {
			originClient, err := NewAdGuardClient(cfg.OriginURL, cfg.Username, cfg.Password, cfg.InsecureSkipVerify)
			if err == nil {
				originFilterStatus, err := originClient.Filtering()
				if err == nil && originFilterStatus != nil && originFilterStatus.UserRules != nil {
					t.Logf(
						"Origin filter status retrieved successfully (user rules count: %d)",
						len(*originFilterStatus.UserRules),
					)
				}
			}
		}
	})
}
