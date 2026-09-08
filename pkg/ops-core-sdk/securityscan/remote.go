package securityscan

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/software"
	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/vulnerability"
)

const remoteMatchPath = "/v1/security/vulnerabilities/match"
const maxRemotePayloadSize = 4 << 20

type remoteMatchRequest struct {
	Inventory software.InventoryResult `json:"inventory"`
}

type remoteMatchResponse struct {
	Findings   []vulnerability.Finding `json:"findings"`
	RuleSource RuleSourceInfo          `json:"rule_source"`
}

// RemoteMatchHandler serves vulnerability matching from a controller-owned rule bundle.
// The handler is transport-agnostic; callers should expose it through HTTPS.
func RemoteMatchHandler(bundle []byte, token string) (http.Handler, error) {
	validated, source, err := ValidateRuleBundle(bundle)
	if err != nil {
		return nil, fmt.Errorf("validate remote rule bundle: %w", err)
	}
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("remote vulnerability token is required")
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != remoteMatchPath {
			http.NotFound(writer, request)
			return
		}
		if request.Method != http.MethodPost {
			writer.Header().Set("Allow", http.MethodPost)
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		expectedAuthorization := []byte("Bearer " + token)
		actualAuthorization := []byte(request.Header.Get("Authorization"))
		if len(actualAuthorization) != len(expectedAuthorization) || subtle.ConstantTimeCompare(actualAuthorization, expectedAuthorization) != 1 {
			http.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		var payload remoteMatchRequest
		decoder := json.NewDecoder(http.MaxBytesReader(writer, request.Body, maxRemotePayloadSize))
		if err := decoder.Decode(&payload); err != nil {
			http.Error(writer, "invalid inventory", http.StatusBadRequest)
			return
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			http.Error(writer, "invalid inventory", http.StatusBadRequest)
			return
		}
		findings := vulnerability.Match(payload.Inventory, validated.Rules)
		response := remoteMatchResponse{Findings: findings, RuleSource: source}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(response); err != nil {
			return
		}
	}), nil
}

func queryRemoteVulnerabilities(ctx context.Context, inventory software.InventoryResult, config *RemoteConfig) ([]vulnerability.Finding, RuleSourceInfo, error) {
	if err := ValidateRemoteConfig(config); err != nil {
		return nil, RuleSourceInfo{}, err
	}
	if strings.TrimSpace(config.Backend) == "external" {
		return queryExternalVulnerabilities(ctx, inventory, config)
	}
	endpoint, _ := url.Parse(strings.TrimSpace(config.URL))
	requestBody, err := json.Marshal(remoteMatchRequest{Inventory: inventory})
	if err != nil {
		return nil, RuleSourceInfo{}, fmt.Errorf("encode remote inventory: %w", err)
	}
	requestContext := ctx
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		requestContext, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}
	request, err := http.NewRequestWithContext(requestContext, http.MethodPost, strings.TrimRight(endpoint.String(), "/")+remoteMatchPath, bytes.NewReader(requestBody))
	if err != nil {
		return nil, RuleSourceInfo{}, fmt.Errorf("create remote vulnerability request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+config.Token)
	request.Header.Set("Content-Type", "application/json")
	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, RuleSourceInfo{}, fmt.Errorf("default HTTP transport is not a *http.Transport")
	}
	transport := baseTransport.Clone()
	if len(config.CA) > 0 {
		pool, poolErr := x509.SystemCertPool()
		if poolErr != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(config.CA) {
			return nil, RuleSourceInfo{}, fmt.Errorf("remote vulnerability CA bundle contains no certificates")
		}
		transport.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS13}
	}
	client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
	if config.Timeout > 0 {
		client.Timeout = config.Timeout
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, RuleSourceInfo{}, fmt.Errorf("query remote vulnerabilities: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, RuleSourceInfo{}, fmt.Errorf("remote vulnerability service returned HTTP %d", response.StatusCode)
	}
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxRemotePayloadSize+1))
	if err != nil {
		return nil, RuleSourceInfo{}, fmt.Errorf("read remote vulnerability response: %w", err)
	}
	if len(responseBody) > maxRemotePayloadSize {
		return nil, RuleSourceInfo{}, fmt.Errorf("remote vulnerability response exceeds %d bytes", maxRemotePayloadSize)
	}
	var payload remoteMatchResponse
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return nil, RuleSourceInfo{}, fmt.Errorf("decode remote vulnerability response: %w", err)
	}
	if config.RuleVersion != "" && payload.RuleSource.RuleVersion != config.RuleVersion {
		return nil, RuleSourceInfo{}, fmt.Errorf("remote rule version mismatch: expected %q, got %q", config.RuleVersion, payload.RuleSource.RuleVersion)
	}
	if config.RuleSHA256 != "" && !strings.EqualFold(config.RuleSHA256, payload.RuleSource.SHA256) {
		return nil, RuleSourceInfo{}, fmt.Errorf("remote rule sha256 mismatch")
	}
	return payload.Findings, payload.RuleSource, nil
}

// ValidateRemoteConfig validates task-scoped controller query settings without
// making a network request.
func ValidateRemoteConfig(config *RemoteConfig) error {
	if config == nil {
		return fmt.Errorf("remote vulnerability configuration is nil")
	}
	endpoint, err := url.Parse(strings.TrimSpace(config.URL))
	if err != nil || endpoint.Scheme != "https" || endpoint.Host == "" {
		return fmt.Errorf("remote vulnerability URL must be an HTTPS URL")
	}
	if endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" || (endpoint.Path != "" && endpoint.Path != "/") {
		return fmt.Errorf("remote vulnerability URL must contain only scheme and host")
	}
	if strings.TrimSpace(config.Token) == "" {
		return fmt.Errorf("remote vulnerability token is required")
	}
	if backend := strings.TrimSpace(config.Backend); backend != "" && backend != "external" && backend != "native" {
		return fmt.Errorf("unsupported remote scanner backend %q", config.Backend)
	}
	if config.Timeout < 0 {
		return fmt.Errorf("remote vulnerability timeout must not be negative")
	}
	if config.RuleSHA256 != "" {
		digest, err := hex.DecodeString(config.RuleSHA256)
		if err != nil || len(digest) != sha256.Size {
			return fmt.Errorf("remote rule sha256 must be a 64-character hexadecimal digest")
		}
	}
	if len(config.CA) > 0 {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(config.CA) {
			return fmt.Errorf("remote vulnerability CA bundle contains no certificates")
		}
	}
	return nil
}
