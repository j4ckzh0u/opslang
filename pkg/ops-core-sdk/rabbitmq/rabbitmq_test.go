package rabbitmq

import (
	"encoding/json"
	"testing"
)

func TestVhostResultValidation(t *testing.T) {
	r := AddVhost("")
	if r.Error == "" {
		t.Error("empty name should return error")
	}
}

func TestUserResultValidation(t *testing.T) {
	r := AddUser("", "pass", "")
	if r.Error == "" {
		t.Error("empty name should return error")
	}
	r = AddUser("test", "", "")
	if r.Error == "" {
		t.Error("empty password should return error")
	}
}

func TestDeleteVhostValidation(t *testing.T) {
	r := DeleteVhost("")
	if r.Error == "" {
		t.Error("empty name should return error")
	}
}

func TestDeleteUserValidation(t *testing.T) {
	r := DeleteUser("")
	if r.Error == "" {
		t.Error("empty name should return error")
	}
}

func TestSetUserTagsValidation(t *testing.T) {
	r := SetUserTags("", "admin")
	if r.Error == "" {
		t.Error("empty name should return error")
	}
}

func TestSetPermissionValidation(t *testing.T) {
	r := SetPermission("", "/", ".*", ".*", ".*")
	if r.Error == "" {
		t.Error("empty user should return error")
	}
	r = SetPermission("guest", "", ".*", ".*", ".*")
	if r.Error == "" {
		t.Error("empty vhost should return error")
	}
}

func TestClearPermissionValidation(t *testing.T) {
	r := ClearPermission("", "/")
	if r.Error == "" {
		t.Error("empty user should return error")
	}
}

func TestSetPolicyValidation(t *testing.T) {
	r := SetPolicy("", "/", ".*", "{}", "all")
	if r.Error == "" {
		t.Error("empty name should return error")
	}
	r = SetPolicy("test", "", ".*", "{}", "all")
	if r.Error == "" {
		t.Error("empty vhost should return error")
	}
}

func TestSetPolicyInvalidJSON(t *testing.T) {
	r := SetPolicy("test", "/", ".*", "not-json", "all")
	if r.Error == "" {
		t.Error("invalid JSON definition should return error")
	}
}

func TestDeletePolicyValidation(t *testing.T) {
	r := DeletePolicy("", "/")
	if r.Error == "" {
		t.Error("empty name should return error")
	}
}

func TestDeclareQueueValidation(t *testing.T) {
	r := DeclareQueue("", "/", "", true, false)
	if r.Error == "" {
		t.Error("empty name should return error")
	}
}

func TestDeleteQueueValidation(t *testing.T) {
	r := DeleteQueue("", "/")
	if r.Error == "" {
		t.Error("empty name should return error")
	}
}

func TestDeclareExchangeValidation(t *testing.T) {
	r := DeclareExchange("", "/", "direct", true, false)
	if r.Error == "" {
		t.Error("empty name should return error")
	}
}

func TestDeleteExchangeValidation(t *testing.T) {
	r := DeleteExchange("", "/")
	if r.Error == "" {
		t.Error("empty name should return error")
	}
}

func TestBindQueueValidation(t *testing.T) {
	r := BindQueue("", "exchange", "/", "")
	if r.Error == "" {
		t.Error("empty queue should return error")
	}
	r = BindQueue("queue", "", "/", "")
	if r.Error == "" {
		t.Error("empty exchange should return error")
	}
}

func TestUnbindQueueValidation(t *testing.T) {
	r := UnbindQueue("", "exchange", "/", "")
	if r.Error == "" {
		t.Error("empty queue should return error")
	}
}

func TestJSONSerialization(t *testing.T) {
	r := VhostResult{Name: "test", Success: true, Changed: true}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out VhostResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if out.Name != "test" || !out.Success || !out.Changed {
		t.Errorf("roundtrip failed: %+v", out)
	}
}

func TestQueueResultJSON(t *testing.T) {
	r := QueueResult{Name: "q1", Vhost: "/", Success: true}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out QueueResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if out.Name != "q1" {
		t.Errorf("expected q1, got %s", out.Name)
	}
}

func TestStatusResultJSON(t *testing.T) {
	r := StatusResult{Node: "rabbit@localhost", Running: true}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out StatusResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !out.Running {
		t.Error("expected running=true")
	}
}
