package securityscan

import "testing"

func TestValidateServiceConfig(t *testing.T) {
	valid := &ServiceConfig{URL: "https://scanner.example", Token: "session-token"}
	if err := ValidateServiceConfig(valid); err != nil {
		t.Fatalf("valid service config rejected: %v", err)
	}
	invalid := []ServiceConfig{
		{URL: "http://scanner.example", Token: "token"},
		{URL: "https://scanner.example", Token: ""},
		{URL: "https://scanner.example?x=1", Token: "token"},
		{URL: "https://scanner.example", Token: "token", Timeout: -1},
	}
	for _, config := range invalid {
		if err := ValidateServiceConfig(&config); err == nil {
			t.Errorf("invalid config accepted: %#v", config)
		}
	}
}

func TestServiceConfigCloneCopiesCA(t *testing.T) {
	config := &ServiceConfig{URL: "https://scanner.example", Token: "token", CA: []byte("ca")}
	clone, err := config.Clone()
	if err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
	clone.CA[0] = 'X'
	if config.CA[0] != 'c' {
		t.Fatal("Clone() must copy CA bytes")
	}
}
