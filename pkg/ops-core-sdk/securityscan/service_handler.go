package securityscan

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
)

const maxServiceRequestSize = 16 << 20

// ScanServicePath returns the stable controller scan endpoint.
func ScanServicePath() string { return serviceScanPath }

// ScanHandler exposes a ScannerBackend through the controller scan protocol.
func ScanHandler(backend ScannerBackend, token string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != serviceScanPath {
			http.NotFound(writer, request)
			return
		}
		if request.Method != http.MethodPost {
			writer.Header().Set("Allow", http.MethodPost)
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		expected := []byte("Bearer " + token)
		actual := []byte(request.Header.Get("Authorization"))
		if strings.TrimSpace(token) == "" || len(actual) != len(expected) || subtle.ConstantTimeCompare(actual, expected) != 1 {
			http.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		if backend == nil {
			http.Error(writer, "scanner backend unavailable", http.StatusServiceUnavailable)
			return
		}
		var scanRequest ScanRequest
		decoder := json.NewDecoder(http.MaxBytesReader(writer, request.Body, maxServiceRequestSize))
		err := decoder.Decode(&scanRequest)
		if err == nil {
			err = decoder.Decode(&struct{}{})
			if err == io.EOF {
				err = ValidateScanRequest(scanRequest)
			} else if err == nil {
				err = errors.New("multiple JSON values")
			}
		}
		if err != nil {
			status := http.StatusBadRequest
			var sizeError *http.MaxBytesError
			if errors.As(err, &sizeError) {
				status = http.StatusRequestEntityTooLarge
			}
			http.Error(writer, "invalid scan request", status)
			return
		}
		capabilities := backend.Capabilities()
		for _, capability := range scanRequest.Scanners {
			if !slices.Contains(capabilities, capability) {
				http.Error(writer, "unsupported scanner capability", http.StatusBadRequest)
				return
			}
		}
		ctx, cancel := context.WithTimeout(request.Context(), ServiceConfigTimeout(scanRequest.Timeout))
		defer cancel()
		report, err := backend.Scan(ctx, scanRequest)
		if err != nil {
			slog.Error("scanner backend execution failed", "backend", backend.Name())
			http.Error(writer, "scan failed", http.StatusBadGateway)
			return
		}
		switch report.Status {
		case StatusOK, StatusPartial, StatusUnsupported, StatusFailed:
		default:
			http.Error(writer, "invalid scanner report", http.StatusBadGateway)
			return
		}
		report.Finish()
		payload, err := json.Marshal(report)
		if err != nil || len(payload) > maxServiceResponseSize {
			slog.Error("scanner backend report cannot be encoded within response limit")
			http.Error(writer, "invalid scanner report", http.StatusBadGateway)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if _, err := writer.Write(payload); err != nil {
			slog.Error("scanner response write failed")
		}
	})
}
