package cmd

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/bakito/adguardhome-sync/internal/config"
	"github.com/bakito/adguardhome-sync/internal/log"
)

var (
	healthcheckPort               int
	healthcheckURL                string
	healthcheckTimeout            time.Duration
	healthcheckInsecureSkipVerify bool
	healthcheckReady              bool
)

// healthcheckCmd represents the healthcheck command.
var healthcheckCmd = &cobra.Command{
	Use:     "healthcheck",
	Aliases: []string{"health"},
	Short:   "Check the health of the running application",
	Long:    `Sends an HTTP request to the /livez (or /readyz) endpoint of the running adguardhome-sync instance to verify responsiveness.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		logger = log.GetLogger("healthcheck")
		cfg, err := config.Get(cfgFile, cmd.Flags())
		if err != nil {
			logger.Error(err)
			return err
		}

		targetURL := healthcheckURL
		if targetURL == "" {
			port := cfg.Get().API.Port
			if healthcheckPort != 0 {
				port = healthcheckPort
			}
			if port == 0 {
				err := errors.New("API port is 0 (API is disabled), health check cannot be performed")
				logger.Error(err)
				return err
			}
			scheme := "http"
			if cfg.Get().API.TLS.Enabled() {
				scheme = "https"
			}
			endpoint := "livez"
			if healthcheckReady {
				endpoint = "readyz"
			}
			targetURL = fmt.Sprintf("%s://localhost:%d/%s", scheme, port, endpoint)
		}

		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, targetURL, http.NoBody)
		if err != nil {
			logger.Error(err)
			return err
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: healthcheckInsecureSkipVerify || cfg.Get().API.TLS.Enabled(), // #nosec G402
			},
		}

		httpClient := &http.Client{
			Timeout:   healthcheckTimeout,
			Transport: tr,
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			logger.With("url", targetURL, "error", err).Error("Health check request failed")
			return err
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("health check failed with status %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
			logger.With("url", targetURL, "status", resp.StatusCode).Error(err)
			return err
		}

		logger.With("url", targetURL).Info("Health check successful")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthcheckCmd)
	healthcheckCmd.Flags().
		IntVarP(&healthcheckPort, "port", "p", 0, "Target API port (defaults to configured API port or 8080)")
	healthcheckCmd.Flags().StringVarP(&healthcheckURL, "url", "u", "", "Target healthcheck URL (overrides port and scheme)")
	healthcheckCmd.Flags().DurationVarP(&healthcheckTimeout, "timeout", "t", 5*time.Second, "Health check request timeout")
	healthcheckCmd.Flags().
		BoolVarP(&healthcheckInsecureSkipVerify, "insecure-skip-verify", "k", false, "Skip TLS certificate verification")
	healthcheckCmd.Flags().
		BoolVarP(&healthcheckReady, "ready", "r", false, "Check readiness (/readyz) instead of liveness (/livez)")
	healthcheckCmd.Flags().Int(config.FlagAPIPort, 0, "Sync API Port")
}
