package securityscan

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type handlerTestBackend struct {
	err error
}

func (b handlerTestBackend) Name() string               { return "test" }
func (b handlerTestBackend) Capabilities() []Capability { return []Capability{CapabilityVulnerability} }
func (b handlerTestBackend) Scan(context.Context, ScanRequest) (ScanReport, error) {
	if b.err != nil {
		return ScanReport{}, b.err
	}
	return ScanReport{Status: StatusOK}, nil
}

func TestScanHandlerValidation(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		path    string
		token   string
		body    string
		backend ScannerBackend
		want    int
	}{
		{"wrong path", http.MethodPost, "/health", "token", `{}`, handlerTestBackend{}, http.StatusNotFound},
		{"wrong method", http.MethodGet, serviceScanPath, "token", `{}`, handlerTestBackend{}, http.StatusMethodNotAllowed},
		{"missing auth", http.MethodPost, serviceScanPath, "token", `{}`, handlerTestBackend{}, http.StatusUnauthorized},
		{"empty token", http.MethodPost, serviceScanPath, "", `{}`, handlerTestBackend{}, http.StatusUnauthorized},
		{"nil backend", http.MethodPost, serviceScanPath, "token", `{"target":{"id":"host"},"scanners":["vulnerability"]}`, nil, http.StatusServiceUnavailable},
		{"invalid json", http.MethodPost, serviceScanPath, "token", `{`, handlerTestBackend{}, http.StatusBadRequest},
		{"missing target", http.MethodPost, serviceScanPath, "token", `{"scanners":["vulnerability"]}`, handlerTestBackend{}, http.StatusBadRequest},
		{"unsupported capability", http.MethodPost, serviceScanPath, "token", `{"target":{"id":"host"},"scanners":["secret"]}`, handlerTestBackend{}, http.StatusBadRequest},
		{"backend failure", http.MethodPost, serviceScanPath, "token", `{"target":{"id":"host"},"scanners":["vulnerability"]}`, handlerTestBackend{err: errors.New("backend failure")}, http.StatusBadGateway},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.name != "missing auth" {
				request.Header.Set("Authorization", "Bearer "+test.token)
			}
			response := httptest.NewRecorder()
			ScanHandler(test.backend, test.token).ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status = %d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestScanHandlerSuccess(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, serviceScanPath, strings.NewReader(`{"target":{"id":"host"},"scanners":["vulnerability"]}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	ScanHandler(handlerTestBackend{}, "token").ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"status":"ok"`) {
		t.Fatalf("response = %q", response.Body.String())
	}
}
