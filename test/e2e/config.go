//go:build e2e

package e2e

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Target string

const (
	TargetK8s Target = "k8s"
)

// Config holds configuration parameters for the E2E test run.
type Config struct {
	Target             Target
	ExternalSync       bool
	Protocol           string
	Mode               string
	SyncURL            string
	SyncPod            string
	OriginURL          string
	OriginPod          string
	ReplicaURLs        []string
	ReplicaVersions    []string
	ReplicaPods        []string
	Username           string
	Password           string
	InsecureSkipVerify bool
	Namespace          string
	Timeout            time.Duration
	StepSummaryFile    string
	LocalBinary        string
	LocalConfigFile    string
}

// LoadConfig reads the E2E test configuration from environment variables with sensible defaults.
func LoadConfig() *Config {
	externalSync := os.Getenv("E2E_EXTERNAL_SYNC") == "true" ||
		os.Getenv("E2E_TARGET") == "local" ||
		os.Getenv("E2E_TARGET") == "external"

	target := TargetK8s
	mode := cmp.Or(os.Getenv("E2E_MODE"), "env")
	protocol := os.Getenv("E2E_PROTOCOL")
	if protocol == "" {
		if !externalSync && mode == "env" {
			protocol = "https"
		} else {
			protocol = "http"
		}
	}
	namespace := cmp.Or(os.Getenv("E2E_NAMESPACE"), "agh-e2e")
	username := cmp.Or(os.Getenv("E2E_USERNAME"), "username")
	password := cmp.Or(os.Getenv("E2E_PASSWORD"), "password")
	syncPod := cmp.Or(os.Getenv("E2E_SYNC_POD"), "adguardhome-sync")
	originPod := cmp.Or(os.Getenv("E2E_ORIGIN_POD"), "adguardhome-origin")

	syncURL := os.Getenv("E2E_SYNC_URL")
	if syncURL == "" {
		syncURL = protocol + "://localhost:9090"
	}

	originURL := os.Getenv("E2E_ORIGIN_URL")
	if originURL == "" {
		originURL = "http://localhost:3000"
	}

	replicaVersionsStr := os.Getenv("E2E_REPLICA_VERSIONS")
	var replicaVersions []string
	if replicaVersionsStr != "" {
		replicaVersions = splitAndTrim(replicaVersionsStr)
	} else {
		replicaVersions = readReplicaVersionsFromValues()
		if len(replicaVersions) == 0 {
			replicaVersions = []string{"v0.107.68", "v0.107.79", "latest"}
		}
	}

	var replicaURLs []string
	if rawURLs := os.Getenv("E2E_REPLICA_URLS"); rawURLs != "" {
		replicaURLs = splitAndTrim(rawURLs)
	} else {
		// Default local ports: 9091, 9092, 9093 for replicas
		for i := range replicaVersions {
			replicaURLs = append(replicaURLs, fmt.Sprintf("http://localhost:%d", 9091+i))
		}
	}

	var replicaPods []string
	if rawPods := os.Getenv("E2E_REPLICA_PODS"); rawPods != "" {
		replicaPods = splitAndTrim(rawPods)
	} else {
		for _, v := range replicaVersions {
			cleanVersion := strings.ReplaceAll(v, ".", "-")
			replicaPods = append(replicaPods, "adguardhome-replica-"+cleanVersion)
		}
	}

	isv := true
	if isvStr := os.Getenv("E2E_INSECURE_SKIP_VERIFY"); isvStr != "" {
		if parsed, err := strconv.ParseBool(isvStr); err == nil {
			isv = parsed
		}
	}

	timeout := 60 * time.Second
	if tStr := os.Getenv("E2E_TIMEOUT"); tStr != "" {
		if d, err := time.ParseDuration(tStr); err == nil {
			timeout = d
		}
	}

	localBinary := cmp.Or(os.Getenv("E2E_LOCAL_BINARY"), "./dist/adguardhome-sync")
	localConfigFile := cmp.Or(os.Getenv("E2E_LOCAL_CONFIG_FILE"), "testdata/e2e/local-config.yaml")
	if !filepath.IsAbs(localConfigFile) {
		if _, err := os.Stat(localConfigFile); os.IsNotExist(err) {
			if _, err := os.Stat(filepath.Join("..", "..", localConfigFile)); err == nil {
				if abs, err := filepath.Abs(filepath.Join("..", "..", localConfigFile)); err == nil {
					localConfigFile = abs
				}
			}
		} else {
			if abs, err := filepath.Abs(localConfigFile); err == nil {
				localConfigFile = abs
			}
		}
	}

	return &Config{
		Target:             target,
		ExternalSync:       externalSync,
		Protocol:           protocol,
		Mode:               mode,
		SyncURL:            syncURL,
		SyncPod:            syncPod,
		OriginURL:          originURL,
		OriginPod:          originPod,
		ReplicaURLs:        replicaURLs,
		ReplicaVersions:    replicaVersions,
		ReplicaPods:        replicaPods,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: isv,
		Namespace:          namespace,
		Timeout:            timeout,
		StepSummaryFile:    os.Getenv("GITHUB_STEP_SUMMARY"),
		LocalBinary:        localBinary,
		LocalConfigFile:    localConfigFile,
	}
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

type e2eReplicaValues struct {
	Versions []string `yaml:"versions"`
}

type e2eValues struct {
	Replica e2eReplicaValues `yaml:"replica"`
}

func readReplicaVersionsFromValues() []string {
	paths := []string{
		"testdata/e2e/values.yaml",
		"../../testdata/e2e/values.yaml",
	}
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			var values e2eValues
			if err := yaml.Unmarshal(data, &values); err == nil && len(values.Replica.Versions) > 0 {
				return values.Replica.Versions
			}
		}
	}
	return nil
}
