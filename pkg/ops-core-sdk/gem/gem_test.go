package gem

import (
	"encoding/json"
	"testing"
)

func TestInstallValidation(t *testing.T) {
	_, err := Install("", "", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUninstallValidation(t *testing.T) {
	_, err := Uninstall("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestInfoValidation(t *testing.T) {
	_, err := Info("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestActionResultJSON(t *testing.T) {
	r := ActionResult{Name: "rails", Success: true, Changed: true, Duration: 100}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ActionResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Success || decoded.Name != "rails" {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestGemInfoJSON(t *testing.T) {
	info := GemInfo{Name: "bundler", Version: "2.4.0", Default: false}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GemInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Name != "bundler" || decoded.Version != "2.4.0" {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestInfoResultJSON(t *testing.T) {
	r := InfoResult{Name: "rails", Found: true, Info: GemInfo{Name: "rails", Version: "7.0.0"}}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded InfoResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Found || decoded.Info.Version != "7.0.0" {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestListResultJSON(t *testing.T) {
	r := ListResult{Gems: []GemInfo{{Name: "rake", Version: "13.0"}}, Count: 1}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ListResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Count != 1 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestVersionResultJSON(t *testing.T) {
	r := VersionResult{Version: "3.4.0", Raw: "3.4.0"}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded VersionResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Version != "3.4.0" {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestParseGemLine(t *testing.T) {
	res, err := parseGemLine("rails (7.0.0, 6.1.0)", "rails")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Found || res.Info.Version != "7.0.0" {
		t.Fatalf("unexpected: %+v", res)
	}
}

func TestParseGemLineNoVersion(t *testing.T) {
	res, err := parseGemLine("rails", "rails")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Found {
		t.Fatal("expected found")
	}
}
