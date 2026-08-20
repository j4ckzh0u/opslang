package kubernetes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsInlineYAML(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{"inline with newline", "apiVersion: v1\nkind: Pod", true},
		{"inline with apiVersion", "apiVersion: v1", true},
		{"inline with kind", "kind: Pod", true},
		{"file path", "/tmp/manifest.yaml", false},
		{"relative path", "manifest.yaml", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isInlineYAML(tc.input); got != tc.expect {
				t.Errorf("isInlineYAML(%q) = %v, want %v", tc.input, got, tc.expect)
			}
		})
	}
}

func TestManifestArg_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte("apiVersion: v1"), 0644); err != nil {
		t.Fatal(err)
	}
	got, cleanup, err := manifestArg(path)
	if err != nil {
		t.Fatalf("manifestArg file: %v", err)
	}
	defer cleanup()
	if !strings.HasSuffix(got, "test.yaml") {
		t.Errorf("expected absolute path ending in test.yaml, got %s", got)
	}
}

func TestManifestArg_Inline(t *testing.T) {
	yaml := "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: test"
	got, cleanup, err := manifestArg(yaml)
	if err != nil {
		t.Fatalf("manifestArg inline: %v", err)
	}
	defer cleanup()
	if !strings.Contains(got, "ops-k8s-") {
		t.Errorf("expected temp file path, got %s", got)
	}
	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}
	if string(data) != yaml {
		t.Errorf("temp file content mismatch: got %q, want %q", string(data), yaml)
	}
}

func TestManifestArg_MissingFile(t *testing.T) {
	_, _, err := manifestArg("/nonexistent/manifest.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestKubectl_NotFound(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)
	_, err := kubectl()
	if err == nil {
		t.Error("expected error when kubectl not in PATH")
	}
	if !strings.Contains(err.Error(), "kubectl not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestToStringMap(t *testing.T) {
	m := map[string]interface{}{
		"key1": "val1",
		"key2": "val2",
		"key3": 42, // non-string should be skipped
	}
	got := toStringMap(m)
	if got["key1"] != "val1" || got["key2"] != "val2" {
		t.Errorf("toStringMap: unexpected result: %v", got)
	}
	if _, exists := got["key3"]; exists {
		t.Error("toStringMap: non-string value should be skipped")
	}
}

func TestParseResource(t *testing.T) {
	raw := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "my-cm",
			"namespace": "default",
			"labels": map[string]interface{}{
				"app": "test",
			},
		},
	}
	r := parseResource(raw)
	if r.Name != "my-cm" {
		t.Errorf("expected name my-cm, got %s", r.Name)
	}
	if r.Namespace != "default" {
		t.Errorf("expected namespace default, got %s", r.Namespace)
	}
	if r.Kind != "ConfigMap" {
		t.Errorf("expected kind ConfigMap, got %s", r.Kind)
	}
	if r.APIVersion != "v1" {
		t.Errorf("expected apiVersion v1, got %s", r.APIVersion)
	}
	if r.Labels["app"] != "test" {
		t.Errorf("expected label app=test, got %v", r.Labels)
	}
}

func TestParsePod(t *testing.T) {
	raw := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "nginx-abc",
			"namespace": "default",
		},
		"status": map[string]interface{}{
			"phase": "Running",
			"podIP": "10.0.0.1",
		},
		"spec": map[string]interface{}{
			"nodeName": "node-1",
			"containers": []interface{}{
				map[string]interface{}{"name": "nginx"},
			},
		},
	}
	p := parsePod(raw)
	if p.Name != "nginx-abc" {
		t.Errorf("expected pod name nginx-abc, got %s", p.Name)
	}
	if p.Status != "Running" {
		t.Errorf("expected status Running, got %s", p.Status)
	}
	if p.IP != "10.0.0.1" {
		t.Errorf("expected IP 10.0.0.1, got %s", p.IP)
	}
	if p.Node != "node-1" {
		t.Errorf("expected node node-1, got %s", p.Node)
	}
	if len(p.Containers) != 1 || p.Containers[0] != "nginx" {
		t.Errorf("expected containers [nginx], got %v", p.Containers)
	}
}

func TestParseService(t *testing.T) {
	raw := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "my-svc",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"type":      "ClusterIP",
			"clusterIP": "10.96.0.1",
			"selector": map[string]interface{}{
				"app": "web",
			},
		},
	}
	s := parseService(raw)
	if s.Name != "my-svc" {
		t.Errorf("expected name my-svc, got %s", s.Name)
	}
	if s.Type != "ClusterIP" {
		t.Errorf("expected type ClusterIP, got %s", s.Type)
	}
	if s.ClusterIP != "10.96.0.1" {
		t.Errorf("expected clusterIP 10.96.0.1, got %s", s.ClusterIP)
	}
	if s.Selector["app"] != "web" {
		t.Errorf("expected selector app=web, got %v", s.Selector)
	}
}

func TestParseDeployment(t *testing.T) {
	raw := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "nginx-deploy",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"replicas": float64(3),
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "nginx",
							"image": "nginx:1.21",
						},
					},
				},
			},
		},
		"status": map[string]interface{}{
			"readyReplicas":     float64(3),
			"availableReplicas": float64(3),
			"updatedReplicas":   float64(3),
		},
	}
	d := parseDeployment(raw)
	if d.Name != "nginx-deploy" {
		t.Errorf("expected name nginx-deploy, got %s", d.Name)
	}
	if d.Replicas != 3 {
		t.Errorf("expected replicas 3, got %d", d.Replicas)
	}
	if d.ReadyReplicas != 3 {
		t.Errorf("expected readyReplicas 3, got %d", d.ReadyReplicas)
	}
	if len(d.Containers) != 1 || d.Containers[0] != "nginx" {
		t.Errorf("expected containers [nginx], got %v", d.Containers)
	}
	if len(d.Images) != 1 || d.Images[0] != "nginx:1.21" {
		t.Errorf("expected images [nginx:1.21], got %v", d.Images)
	}
	if d.Ready != "3/3" {
		t.Errorf("expected ready 3/3, got %s", d.Ready)
	}
}

// TestApply_NoKubectl verifies Apply fails gracefully when kubectl is missing.
func TestApply_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	yaml := "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: test"
	_, err := Apply(yaml, "default", false)
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
	if !strings.Contains(err.Error(), "kubectl not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDelete_NoKubectl verifies Delete fails gracefully when kubectl is missing.
func TestDelete_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	yaml := "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: test"
	_, err := Delete(yaml, "default")
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestExec_NoKubectl verifies Exec fails gracefully when kubectl is missing.
func TestExec_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := Exec("mypod", "ls", "default", "")
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestScale_NoKubectl verifies Scale fails gracefully when kubectl is missing.
func TestScale_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := Scale("mydeploy", 3, "default")
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestGet_NoKubectl verifies Get fails gracefully when kubectl is missing.
func TestGet_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := Get("pod", "mypod", "default")
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestList_NoKubectl verifies List fails gracefully when kubectl is missing.
func TestList_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := List("pods", "default", "")
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestCreateNamespace_NoKubectl verifies CreateNamespace fails gracefully.
func TestCreateNamespace_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := CreateNamespace("test")
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestDeleteNamespace_NoKubectl verifies DeleteNamespace fails gracefully.
func TestDeleteNamespace_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := DeleteNamespace("test")
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestGetPods_NoKubectl verifies GetPods fails gracefully.
func TestGetPods_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := GetPods("default", "")
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestGetServices_NoKubectl verifies GetServices fails gracefully.
func TestGetServices_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := GetServices("default")
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestGetDeployments_NoKubectl verifies GetDeployments fails gracefully.
func TestGetDeployments_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := GetDeployments("default")
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestLogs_NoKubectl verifies Logs fails gracefully.
func TestLogs_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := Logs("mypod", "default", "", 10)
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestRolloutStatus_NoKubectl verifies RolloutStatus fails gracefully.
func TestRolloutStatus_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := RolloutStatus("mydeploy", "default")
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestWaitReady_NoKubectl verifies WaitReady fails gracefully.
func TestWaitReady_NoKubectl(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := WaitReady("deployment", "mydeploy", "default", 30)
	if err == nil {
		t.Error("expected error when kubectl missing")
	}
}

// TestResultJSON verifies JSON serialization of Result.
func TestResultJSON(t *testing.T) {
	r := Result{Changed: true, Status: "applied", Message: "namespace/test created"}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, `"changed":true`) {
		t.Errorf("expected changed:true in JSON, got %s", s)
	}
	if !strings.Contains(s, `"status":"applied"`) {
		t.Errorf("expected status:applied in JSON, got %s", s)
	}
}

// TestExecResultJSON verifies JSON serialization of ExecResult.
func TestExecResultJSON(t *testing.T) {
	r := ExecResult{Stdout: "hello", Stderr: "", ExitCode: 0, Pod: "mypod", Container: "main"}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, `"pod":"mypod"`) {
		t.Errorf("expected pod:mypod in JSON, got %s", s)
	}
}

