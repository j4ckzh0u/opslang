package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/securityscan"
	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/software"
	"github.com/spf13/cobra"
)

var (
	scanPath        string
	scanInventory   string
	scanRules       string
	scanScanners    string
	scanFormat      string
	scanSeverity    string
	scanMaxSize     int64
	scanMaxDepth    int
	scanIgnoreIDs   []string
	scanIgnorePkgs  []string
	scanIgnorePaths []string
)

var errScanGate = errors.New("scan findings met the configured severity threshold")

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan a host inventory or filesystem",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScanCommand(cmd)
	},
}

func init() {
	scanCmd.Flags().StringVar(&scanPath, "path", "", "Filesystem directory to scan")
	scanCmd.Flags().StringVar(&scanInventory, "inventory", "", "Software inventory JSON file")
	scanCmd.Flags().StringVar(&scanRules, "rules", "", "Offline vulnerability rule bundle JSON file")
	scanCmd.Flags().StringVar(&scanScanners, "scanners", "sbom", "Comma-separated scanners: vuln,sbom")
	scanCmd.Flags().StringVar(&scanFormat, "format", "json", "Output format: json, cyclonedx, or spdx")
	scanCmd.Flags().StringVar(&scanSeverity, "severity", "", "Minimum vulnerability severity")
	scanCmd.Flags().Int64Var(&scanMaxSize, "max-file-size", 0, "Maximum manifest file size in bytes")
	scanCmd.Flags().IntVar(&scanMaxDepth, "max-depth", 0, "Maximum filesystem scan depth")
	scanCmd.Flags().StringSliceVar(&scanIgnoreIDs, "ignore-id", nil, "Vulnerability IDs to ignore")
	scanCmd.Flags().StringSliceVar(&scanIgnorePkgs, "ignore-package", nil, "Package names to ignore")
	scanCmd.Flags().StringSliceVar(&scanIgnorePaths, "ignore-path", nil, "Target-relative paths to ignore")
}

func runScanCommand(cmd *cobra.Command) error {
	if (scanPath == "") == (scanInventory == "") {
		return fmt.Errorf("exactly one of --path or --inventory is required")
	}
	scanners, err := parseScanScanners(scanScanners)
	if err != nil {
		return err
	}
	options := securityscan.Options{Scanners: scanners, Severity: scanSeverity, Format: scanFormat, MaxFileSize: scanMaxSize, MaxDepth: scanMaxDepth, IgnorePaths: append([]string(nil), scanIgnorePaths...)}
	for _, id := range scanIgnoreIDs {
		options.IgnoreRules = append(options.IgnoreRules, securityscan.IgnoreRule{ID: id})
	}
	for _, packageName := range scanIgnorePkgs {
		options.IgnoreRules = append(options.IgnoreRules, securityscan.IgnoreRule{Package: packageName})
	}
	if err := securityscan.ValidateOptions(options); err != nil {
		return err
	}
	if scanRules != "" {
		options.RuleBundle, err = os.ReadFile(scanRules)
		if err != nil {
			return fmt.Errorf("read rules: %w", err)
		}
	}
	var result securityscan.ScanResult
	if scanPath != "" {
		result, err = securityscan.ScanFilesystem(context.Background(), scanPath, options)
	} else {
		var inventory software.InventoryResult
		data, readErr := os.ReadFile(scanInventory)
		if readErr != nil {
			return fmt.Errorf("read inventory: %w", readErr)
		}
		if err := json.Unmarshal(data, &inventory); err != nil {
			return fmt.Errorf("decode inventory: %w", err)
		}
		result, err = securityscan.ScanInventory(inventory, options)
	}
	if err != nil {
		return err
	}
	output := interface{}(result)
	if scanFormat == "cyclonedx" || scanFormat == "spdx" {
		output, err = securityscan.GenerateSBOM(result.Components, scanFormat)
		if err != nil {
			return err
		}
	}
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return err
	}
	if result.Status == securityscan.StatusFailed {
		return errScanGate
	}
	return nil
}

func parseScanScanners(value string) ([]securityscan.Scanner, error) {
	parts := strings.Split(value, ",")
	scanners := make([]securityscan.Scanner, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		scanners = append(scanners, securityscan.Scanner(part))
	}
	if err := securityscan.ValidateScanners(scanners); err != nil {
		return nil, err
	}
	return scanners, nil
}
