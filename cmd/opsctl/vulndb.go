package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/securityscan"
	"github.com/spf13/cobra"
)

var (
	vulnDBRules    string
	vulnDBListen   string
	vulnDBCert     string
	vulnDBKey      string
	vulnDBToken    string
	vulnDBShutdown time.Duration
)

var vulnDBCmd = &cobra.Command{Use: "vulndb", Short: "Manage the controller vulnerability query service"}

var vulnDBServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve controller-owned vulnerability matching over HTTPS",
	RunE: func(cmd *cobra.Command, args []string) error {
		bundle, err := os.ReadFile(vulnDBRules)
		if err != nil {
			return fmt.Errorf("read vulnerability rules: %w", err)
		}
		handler, err := securityscan.RemoteMatchHandler(bundle, vulnDBToken)
		if err != nil {
			return err
		}
		backend, err := securityscan.NewRuleBundleBackend(bundle)
		if err != nil {
			return err
		}
		mux := http.NewServeMux()
		mux.Handle(securityscan.ScanServicePath(), securityscan.ScanHandler(backend, vulnDBToken))
		mux.Handle("/v1/security/vulnerabilities/match", handler)
		server := &http.Server{
			Addr:              vulnDBListen,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       30 * time.Second,
		}
		if vulnDBShutdown > 0 {
			go func() {
				<-time.After(vulnDBShutdown)
				if shutdownErr := server.Shutdown(context.Background()); shutdownErr != nil {
					fmt.Fprintf(os.Stderr, "vulnerability service shutdown failed: %v\n", shutdownErr)
				}
			}()
		}
		if err := server.ListenAndServeTLS(vulnDBCert, vulnDBKey); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve vulnerability database: %w", err)
		}
		return nil
	},
}

func init() {
	vulnDBServeCmd.Flags().StringVar(&vulnDBRules, "rules", "", "Controller-owned rule bundle JSON file")
	vulnDBServeCmd.Flags().StringVar(&vulnDBListen, "listen", "127.0.0.1:8443", "HTTPS listen address")
	vulnDBServeCmd.Flags().StringVar(&vulnDBCert, "cert", "", "TLS certificate file")
	vulnDBServeCmd.Flags().StringVar(&vulnDBKey, "key", "", "TLS private key file")
	vulnDBServeCmd.Flags().StringVar(&vulnDBToken, "token", "", "Task-scoped bearer token")
	vulnDBServeCmd.Flags().DurationVar(&vulnDBShutdown, "shutdown-after", 0, "Stop after this duration; intended for tests")
	_ = vulnDBServeCmd.MarkFlagRequired("rules")
	_ = vulnDBServeCmd.MarkFlagRequired("cert")
	_ = vulnDBServeCmd.MarkFlagRequired("key")
	_ = vulnDBServeCmd.MarkFlagRequired("token")
	vulnDBCmd.AddCommand(vulnDBServeCmd)
	rootCmd.AddCommand(vulnDBCmd)
}

func loadRemoteVulnerabilityConfig(endpoint, token, ruleVersion, ruleSHA256, caPath string, timeout time.Duration) (*securityscan.RemoteConfig, error) {
	if !hasRemoteVulnerabilitySettings(endpoint, token, ruleVersion, ruleSHA256, caPath) {
		return nil, nil
	}
	config := &securityscan.RemoteConfig{
		URL:         endpoint,
		Token:       token,
		RuleVersion: ruleVersion,
		RuleSHA256:  ruleSHA256,
		Timeout:     timeout,
	}
	if caPath != "" {
		ca, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("read vulnerability database CA: %w", err)
		}
		config.CA = ca
	}
	if err := securityscan.ValidateRemoteConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}

func hasRemoteVulnerabilitySettings(endpoint, token, ruleVersion, ruleSHA256, caPath string) bool {
	return endpoint != "" || token != "" || ruleVersion != "" || ruleSHA256 != "" || caPath != ""
}
