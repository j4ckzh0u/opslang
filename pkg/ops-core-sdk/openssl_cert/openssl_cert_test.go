package openssl_cert

import (
	"encoding/json"
	"testing"
)

func TestCreateCSRValidation(t *testing.T) {
	_, err := CreateCSR("", "", "", 0)
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestGenerateSelfSignedValidation(t *testing.T) {
	_, err := GenerateSelfSigned("", "", "", 0, 0)
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestInspectValidation(t *testing.T) {
	_, err := Inspect("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestInspectNonExistent(t *testing.T) {
	res, err := Inspect("/nonexistent/cert.pem")
	if err != nil {
		t.Fatal(err)
	}
	if res.Exists {
		t.Fatal("expected exists=false")
	}
}

func TestVerifyValidation(t *testing.T) {
	_, err := Verify("", "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestCheckExpiryValidation(t *testing.T) {
	_, err := CheckExpiry("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestConvertFormatValidation(t *testing.T) {
	_, err := ConvertFormat("", "", "")
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
}

func TestCertInfoJSON(t *testing.T) {
	info := CertInfo{Subject: "CN=test", Issuer: "CN=test", SelfSigned: true}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	var decoded CertInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.SelfSigned || decoded.Subject != "CN=test" {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestCSRResultJSON(t *testing.T) {
	r := CSRResult{CSRPath: "/tmp/test.csr", KeyPath: "/tmp/test.key", Success: true, Duration: 100}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded CSRResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Success || decoded.Duration != 100 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestSelfSignedResultJSON(t *testing.T) {
	r := SelfSignedResult{CertPath: "/tmp/cert.pem", KeyPath: "/tmp/key.pem", Success: true}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded SelfSignedResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Success {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestInspectResultJSON(t *testing.T) {
	r := InspectResult{Exists: true, Info: CertInfo{Subject: "CN=test"}}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded InspectResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Exists {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestVerifyResultJSON(t *testing.T) {
	r := VerifyResult{Valid: false, Errors: []string{"expired"}}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded VerifyResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Valid {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestExpiryResultJSON(t *testing.T) {
	r := ExpiryResult{Path: "/tmp/cert.pem", DaysLeft: 30, Expired: false}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ExpiryResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Expired || decoded.DaysLeft != 30 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}
