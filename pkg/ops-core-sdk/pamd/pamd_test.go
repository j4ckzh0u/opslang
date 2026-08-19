package pamd

import (
	"encoding/json"
	"testing"
)

func TestGetEmptyService(t *testing.T) {
	_, err := Get("")
	if err == nil {
		t.Fatal("expected error for empty service")
	}
}

func TestGetNonExistent(t *testing.T) {
	res, err := Get("nonexistent_service_xyz")
	if err != nil {
		t.Fatal(err)
	}
	if res.Exists {
		t.Fatal("expected exists=false")
	}
}

func TestAddRuleValidation(t *testing.T) {
	_, err := AddRule("", "", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestRemoveRuleValidation(t *testing.T) {
	_, err := RemoveRule("", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestModifyRuleValidation(t *testing.T) {
	_, err := ModifyRule("", "", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestValidateValidation(t *testing.T) {
	_, err := Validate("")
	if err == nil {
		t.Fatal("expected error for empty service")
	}
}

func TestBackupValidation(t *testing.T) {
	_, err := Backup("", "")
	if err == nil {
		t.Fatal("expected error for empty service")
	}
}

func TestParseRules(t *testing.T) {
	content := `# comment
auth	required	pam_unix.so
account	sufficient	pam_permit.so	arg1 arg2
`
	rules := parseRules(content)
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].Type != "auth" || rules[0].Module != "pam_unix.so" {
		t.Fatalf("unexpected rule 0: %+v", rules[0])
	}
	if rules[1].Args != "arg1 arg2" {
		t.Fatalf("expected args 'arg1 arg2', got %q", rules[1].Args)
	}
}

func TestServiceResultJSON(t *testing.T) {
	r := ServiceResult{Service: "sshd", Exists: true, Rules: []Rule{{Type: "auth", Module: "pam_unix.so"}}}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ServiceResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Service != "sshd" || !decoded.Exists {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestActionResultJSON(t *testing.T) {
	r := ActionResult{Service: "sshd", Success: true, Changed: true, Duration: 5}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ActionResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Success || !decoded.Changed {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestValidateResultJSON(t *testing.T) {
	r := ValidateResult{Service: "sshd", Valid: false, Errors: []string{"bad line"}}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ValidateResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Valid || len(decoded.Errors) != 1 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestListResultJSON(t *testing.T) {
	r := ListResult{Services: []string{"sshd", "login"}, Count: 2}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ListResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Count != 2 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}
