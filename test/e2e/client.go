//go:build e2e

package e2e

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bakito/adguardhome-sync/internal/client"
	"github.com/bakito/adguardhome-sync/internal/types"
)

// SyncStatus represents the payload returned by /api/v1/status.
type SyncStatus struct {
	SyncRunning bool            `json:"syncRunning"`
	Origin      ReplicaStatus   `json:"origin"`
	Replicas    []ReplicaStatus `json:"replicas"`
}

// ReplicaStatus represents status details for origin or replica instance.
type ReplicaStatus struct {
	Host              string `json:"host"`
	URL               string `json:"url"`
	Status            string `json:"status"`
	Error             string `json:"error,omitempty"`
	ProtectionEnabled *bool  `json:"protection_enabled"`
}

// HTTPClient creates a configured http.Client with custom TLS settings.
func HTTPClient(insecure bool, timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure, // #nosec G402
			},
		},
	}
}

// GetStatus queries /api/v1/status and parses the response into SyncStatus.
func GetStatus(ctx context.Context, baseURL string, insecure bool) (*SyncStatus, error) {
	apiURL := strings.TrimRight(baseURL, "/") + "/api/v1/status"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	cl := HTTPClient(insecure, 10*time.Second)
	resp, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("status request failed for %s: %w", apiURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var status SyncStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %w", err)
	}
	return &status, nil
}

// GetMetrics scrapes the /metrics endpoint and returns the raw string content.
func GetMetrics(ctx context.Context, baseURL string, insecure bool) (string, error) {
	apiURL := strings.TrimRight(baseURL, "/") + "/metrics"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
	if err != nil {
		return "", err
	}

	cl := HTTPClient(insecure, 10*time.Second)
	resp, err := cl.Do(req)
	if err != nil {
		return "", fmt.Errorf("metrics request failed for %s: %w", apiURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("metrics endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read metrics body: %w", err)
	}
	return string(body), nil
}

// CheckEndpoint makes a GET request to a specific URL and returns the status code.
func CheckEndpoint(ctx context.Context, targetURL string, insecure bool) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, http.NoBody)
	if err != nil {
		return 0, err
	}

	cl := HTTPClient(insecure, 5*time.Second)
	resp, err := cl.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return resp.StatusCode, nil
}

// WaitForSync polls /api/v1/status until SyncRunning is false, returning the final status.
func WaitForSync(ctx context.Context, baseURL string, insecure bool, timeout time.Duration) (*SyncStatus, error) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeoutChan:
			return nil, errors.New("timed out waiting for synchronization to complete")
		case <-ticker.C:
			status, err := GetStatus(ctx, baseURL, insecure)
			if err != nil {
				continue
			}
			if !status.SyncRunning {
				return status, nil
			}
		}
	}
}

// NewAdGuardClient creates an internal Client pointing to an AdGuardHome instance.
func NewAdGuardClient(targetURL, username, password string, insecure bool) (client.Client, error) {
	inst := types.AdGuardInstance{
		Host:               targetURL,
		URL:                targetURL,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: insecure,
	}
	return client.New(inst, 10*time.Second)
}
