package sefcontext

import (
	"testing"
)

func TestListResultJSON(t *testing.T) {
	r := ListResult{
		Contexts: []ContextEntry{
			{SELinuxContext: "unconfined_u:object_r:httpd_log_t:s0", FileSpec: "/var/log/httpd(/.*)?"},
		},
	}
	if len(r.Contexts) != 1 {
		t.Errorf("expected 1 context, got %d", len(r.Contexts))
	}
	if r.Contexts[0].FileSpec != "/var/log/httpd(/.*)?" {
		t.Errorf("unexpected filespec: %s", r.Contexts[0].FileSpec)
	}
}

func TestAddValidation(t *testing.T) {
	_, err := Add("", "httpd_log_t")
	if err == nil {
		t.Error("expected error for empty filespec")
	}

	_, err = Add("/var/log/httpd", "")
	if err == nil {
		t.Error("expected error for empty se_type")
	}
}

func TestModifyValidation(t *testing.T) {
	_, err := Modify("", "httpd_log_t")
	if err == nil {
		t.Error("expected error for empty filespec")
	}

	_, err = Modify("/var/log/httpd", "")
	if err == nil {
		t.Error("expected error for empty se_type")
	}
}

func TestRemoveValidation(t *testing.T) {
	_, err := Remove("")
	if err == nil {
		t.Error("expected error for empty filespec")
	}
}

func TestApplyValidation(t *testing.T) {
	_, err := Apply("", false)
	if err == nil {
		t.Error("expected error for empty filespec")
	}
}

func TestActionResultJSON(t *testing.T) {
	r := ActionResult{Changed: true, Message: "test"}
	if !r.Changed {
		t.Error("expected Changed=true")
	}
}
