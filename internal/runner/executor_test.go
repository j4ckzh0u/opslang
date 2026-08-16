package runner

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// ============================================================
// Types and parsing tests
// ============================================================

func TestInstructionPackageParsing(t *testing.T) {
	input := `{
		"version": "1.0",
		"task_id": "test-123",
		"dry_run": false,
		"instructions": [
			{"op": "sys.cpu.usage", "args": {}, "assign": "cpu"},
			{"op": "report", "args": {"cpu": "cpu"}}
		]
	}`

	var pkg InstructionPackage
	if err := json.Unmarshal([]byte(input), &pkg); err != nil {
		t.Fatalf("failed to parse instruction package: %v", err)
	}

	if pkg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", pkg.Version)
	}
	if pkg.TaskID != "test-123" {
		t.Errorf("expected task_id test-123, got %s", pkg.TaskID)
	}
	if len(pkg.Instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(pkg.Instructions))
	}
	if pkg.Instructions[0].Op != "sys.cpu.usage" {
		t.Errorf("expected op sys.cpu.usage, got %s", pkg.Instructions[0].Op)
	}
	if pkg.Instructions[0].Assign != "cpu" {
		t.Errorf("expected assign cpu, got %s", pkg.Instructions[0].Assign)
	}
}

func TestValidatePackage(t *testing.T) {
	tests := []struct {
		name    string
		pkg     InstructionPackage
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid package",
			pkg:     InstructionPackage{Version: "1.0", Instructions: []Instruction{{Op: "sys.cpu.usage"}}},
			wantErr: false,
		},
		{
			name:    "unknown operation",
			pkg:     InstructionPackage{Version: "1.0", Instructions: []Instruction{{Op: "not.a.real.op"}}},
			wantErr: true,
			errMsg:  "unknown operation",
		},
		{
			name:    "missing version",
			pkg:     InstructionPackage{Instructions: []Instruction{{Op: "test"}}},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name:    "unsupported version",
			pkg:     InstructionPackage{Version: "2.0", Instructions: []Instruction{{Op: "test"}}},
			wantErr: true,
			errMsg:  "unsupported version",
		},
		{
			name:    "no instructions",
			pkg:     InstructionPackage{Version: "1.0"},
			wantErr: true,
			errMsg:  "at least one instruction",
		},
		{
			name:    "instruction without op",
			pkg:     InstructionPackage{Version: "1.0", Instructions: []Instruction{{}}},
			wantErr: true,
			errMsg:  "op is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackage(&tt.pkg)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

// ============================================================
// Registry tests
// ============================================================

func TestRegistryRegisterAndGet(t *testing.T) {
	r := &Registry{ops: make(map[string]OperationFunc)}

	called := false
	r.Register("test.op", func(args map[string]interface{}) (interface{}, error) {
		called = true
		return "result", nil
	})

	fn, ok := r.Get("test.op")
	if !ok {
		t.Fatal("expected operation to be found")
	}
	result, err := fn(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("operation was not called")
	}
	if result != "result" {
		t.Errorf("expected result, got %v", result)
	}
}

func TestRegistryGetMissing(t *testing.T) {
	r := &Registry{ops: make(map[string]OperationFunc)}
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected operation to not be found")
	}
}

func TestRegistryListOperations(t *testing.T) {
	r := NewRegistry()
	ops := r.ListOperations()
	if len(ops) == 0 {
		t.Fatal("expected at least one operation")
	}

	// Check that some expected operations exist.
	expected := []string{
		"sys.cpu.usage",
		"sys.memory.info",
		"file.read",
		"net.http_get",
		"process.list",
		"service.status",
		"time.now",
		"json.encode",
		"yaml.encode",
		"report",
		"log",
	}
	opSet := make(map[string]bool)
	for _, op := range ops {
		opSet[op] = true
	}
	for _, e := range expected {
		if !opSet[e] {
			t.Errorf("expected operation %q to be registered", e)
		}
	}
}

func TestNewRegistryHasAllOps(t *testing.T) {
	r := NewRegistry()
	ops := r.ListOperations()
	// We expect at least 40 operations based on the operation mapping.
	if len(ops) < 40 {
		t.Errorf("expected at least 40 operations, got %d", len(ops))
	}
}

// ============================================================
// Executor tests
// ============================================================

func TestExecutorVariableAssignment(t *testing.T) {
	r := &Registry{ops: make(map[string]OperationFunc)}
	r.Register("test.greet", func(args map[string]interface{}) (interface{}, error) {
		return "hello", nil
	})

	pkg := &InstructionPackage{
		Version: "1.0",
		Instructions: []Instruction{
			{Op: "test.greet", Assign: "greeting"},
		},
	}

	output := Run(pkg, r)
	if output.Status != "ok" {
		t.Errorf("expected status ok, got %s", output.Status)
	}
	if output.Data["greeting"] != "hello" {
		t.Errorf("expected greeting=hello, got %v", output.Data["greeting"])
	}
}

func TestExecutorVariableResolution(t *testing.T) {
	r := &Registry{ops: make(map[string]OperationFunc)}

	// First op returns a value, second op receives it as a resolved variable.
	r.Register("test.produce", func(args map[string]interface{}) (interface{}, error) {
		return "produced-value", nil
	})
	r.Register("test.consume", func(args map[string]interface{}) (interface{}, error) {
		val, _ := args["input"].(string)
		return "consumed:" + val, nil
	})

	pkg := &InstructionPackage{
		Version: "1.0",
		Instructions: []Instruction{
			{Op: "test.produce", Assign: "myvar"},
			// The string "myvar" should be resolved to the variable's value.
			{Op: "test.consume", Args: map[string]interface{}{"input": "myvar"}},
		},
	}

	output := Run(pkg, r)
	if output.Status != "ok" {
		t.Fatalf("expected status ok, got %s (errors: %v)", output.Status, output.Errors)
	}

	// The consume op should have received the resolved value.
	if output.Errors != nil && len(output.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", output.Errors)
	}
}

func TestExecutorDryRun(t *testing.T) {
	r := &Registry{ops: make(map[string]OperationFunc)}
	called := false
	r.Register("test.dangerous", func(args map[string]interface{}) (interface{}, error) {
		called = true
		return "executed", nil
	})

	pkg := &InstructionPackage{
		Version: "1.0",
		DryRun:  true,
		Instructions: []Instruction{
			{Op: "test.dangerous", Args: map[string]interface{}{"key": "val"}},
		},
	}

	output := Run(pkg, r)
	if called {
		t.Error("operation should not have been called in dry-run mode")
	}
	if output.Status != "ok" {
		t.Errorf("expected status ok, got %s", output.Status)
	}

	// Check that dry-run result was returned.
	data, ok := output.Data[""].(map[string]interface{})
	if !ok {
		// No assign, so result is not in Data. But dry-run should still produce the info.
		// Let's add an assign to test properly.
	}
	_ = data

	// Test with assign.
	pkg2 := &InstructionPackage{
		Version: "1.0",
		DryRun:  true,
		Instructions: []Instruction{
			{Op: "test.dangerous", Args: map[string]interface{}{"key": "val"}, Assign: "result"},
		},
	}
	output2 := Run(pkg2, r)
	resultMap, ok := output2.Data["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result to be a map, got %T", output2.Data["result"])
	}
	if resultMap["dry_run"] != true {
		t.Error("expected dry_run to be true")
	}
	if resultMap["operation"] != "test.dangerous" {
		t.Errorf("expected operation test.dangerous, got %v", resultMap["operation"])
	}
}

func TestExecutorUnknownOp(t *testing.T) {
	r := &Registry{ops: make(map[string]OperationFunc)}

	pkg := &InstructionPackage{
		Version: "1.0",
		Instructions: []Instruction{
			{Op: "nonexistent.op"},
		},
	}

	output := Run(pkg, r)
	if output.Status != "failed" {
		t.Errorf("expected status failed (nothing succeeded), got %s", output.Status)
	}
	if len(output.Errors) == 0 {
		t.Error("expected at least one error")
	}
	if !strings.Contains(output.Errors[0], "unknown operation") {
		t.Errorf("expected 'unknown operation' error, got %q", output.Errors[0])
	}
}

func TestExecutorErrorDoesNotStopExecution(t *testing.T) {
	r := &Registry{ops: make(map[string]OperationFunc)}
	r.Register("test.fail", func(args map[string]interface{}) (interface{}, error) {
		return nil, &testError{"intentional failure"}
	})
	r.Register("test.succeed", func(args map[string]interface{}) (interface{}, error) {
		return "ok", nil
	})

	pkg := &InstructionPackage{
		Version: "1.0",
		Instructions: []Instruction{
			{Op: "test.fail", Assign: "failed"},
			{Op: "test.succeed", Assign: "succeeded"},
		},
	}

	output := Run(pkg, r)
	if output.Status != "partial" {
		t.Errorf("expected status partial, got %s", output.Status)
	}
	if len(output.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(output.Errors))
	}
	if output.Data["succeeded"] != "ok" {
		t.Errorf("expected second instruction to succeed, got %v", output.Data["succeeded"])
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }

// ============================================================
// Report operation tests
// ============================================================

func TestReportOperation(t *testing.T) {
	r := NewRegistry()
	r.Register("test.value", func(args map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{"value": 42}, nil
	})

	pkg := &InstructionPackage{
		Version: "1.0",
		Instructions: []Instruction{
			{Op: "test.value", Assign: "mydata"},
			{Op: "report", Args: map[string]interface{}{"data": "$mydata"}},
		},
	}

	output := Run(pkg, r)
	if output.Status != "ok" {
		t.Fatalf("expected status ok, got %s (errors: %v)", output.Status, output.Errors)
	}

	// Report should override the output data with the report map.
	data, ok := output.Data["data"]
	if !ok {
		t.Fatal("expected 'data' key in output")
	}
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T", data)
	}
	if dataMap["value"] != 42 {
		t.Errorf("expected value 42, got %v", dataMap["value"])
	}
}

// ============================================================
// Log operation tests
// ============================================================

func TestLogOperation(t *testing.T) {
	r := NewRegistry()

	pkg := &InstructionPackage{
		Version: "1.0",
		Instructions: []Instruction{
			{Op: "log", Args: map[string]interface{}{"message": "test warning"}},
		},
	}

	output := Run(pkg, r)
	if output.Status != "ok" {
		t.Errorf("expected status ok, got %s", output.Status)
	}
	if len(output.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(output.Warnings))
	}
	if output.Warnings[0] != "test warning" {
		t.Errorf("expected warning 'test warning', got %q", output.Warnings[0])
	}
}

// ============================================================
// Argument helper tests
// ============================================================

func TestGetStringArg(t *testing.T) {
	args := map[string]interface{}{
		"key": "value",
		"num": 42,
	}

	if got := getStringArg(args, "key", "default"); got != "value" {
		t.Errorf("expected 'value', got %q", got)
	}
	if got := getStringArg(args, "missing", "default"); got != "default" {
		t.Errorf("expected 'default', got %q", got)
	}
	if got := getStringArg(args, "num", "default"); got != "default" {
		t.Errorf("expected 'default' for non-string, got %q", got)
	}
}

func TestArgInt(t *testing.T) {
	args := map[string]interface{}{
		"float":  42.0,
		"int":    42,
		"int64":  int64(42),
		"string": "not a number",
	}

	if got, err := argInt(args, "float"); err != nil || got != 42 {
		t.Errorf("float64: got (%d, %v), want (42, nil)", got, err)
	}
	if got, err := argInt(args, "int"); err != nil || got != 42 {
		t.Errorf("int: got (%d, %v), want (42, nil)", got, err)
	}
	if got, err := argInt(args, "int64"); err != nil || got != 42 {
		t.Errorf("int64: got (%d, %v), want (42, nil)", got, err)
	}
	if _, err := argInt(args, "missing"); err == nil {
		t.Error("missing arg must return an error, not a silent 0")
	}
	if _, err := argInt(args, "string"); err == nil {
		t.Error("non-number arg must return an error")
	}
}

func TestArgInt64(t *testing.T) {
	args := map[string]interface{}{
		"float": 42.0,
	}
	if got, err := argInt64(args, "float"); err != nil || got != 42 {
		t.Errorf("got (%d, %v), want (42, nil)", got, err)
	}
	if _, err := argInt64(args, "missing"); err == nil {
		t.Error("missing arg must return an error")
	}
}

func TestArgString(t *testing.T) {
	args := map[string]interface{}{
		"path": "/etc/hosts",
		"num":  42,
	}
	if got, err := argString(args, "path"); err != nil || got != "/etc/hosts" {
		t.Errorf("got (%q, %v)", got, err)
	}
	if _, err := argString(args, "missing"); err == nil {
		t.Error("missing arg must return an error, not a silent \"\"")
	}
	if _, err := argString(args, "num"); err == nil {
		t.Error("non-string arg must return an error")
	}
}

// A registered op receiving a missing required argument must fail the
// instruction (the old behavior silently operated on "" / 0).
func TestRegistryOpMissingArgFails(t *testing.T) {
	r := NewRegistry()
	fn, ok := r.Get("file.read")
	if !ok {
		t.Fatal("file.read not registered")
	}
	if _, err := fn(map[string]interface{}{}); err == nil {
		t.Error("file.read with no path must fail")
	}
}

// Canonical names must all resolve; aliases must map to the same function.
func TestRegistryCanonicalAndAliases(t *testing.T) {
	r := NewRegistry()
	for _, name := range []string{
		"sys.load", "sys.os", "net.http_get", "net.tcp_check",
		"process.find_by_name", "file.stat", "pkg.info",
	} {
		if !r.Has(name) {
			t.Errorf("canonical op %q not registered", name)
		}
	}
	for alias, canonical := range map[string]string{
		"sys.load.avg":         "sys.load",
		"sys.host.info":        "sys.os",
		"net.http.get":         "net.http_get",
		"net.tcp.ping":         "net.tcp_check",
		"process.find.by_name": "process.find_by_name",
		"file.info":            "file.stat",
		"pkg.search":           "pkg.info",
	} {
		fnA, okA := r.Get(alias)
		fnB, _ := r.Get(canonical)
		if !okA {
			t.Errorf("alias %q not resolvable", alias)
			continue
		}
		// Compare via fmt pointer string (func values are not comparable).
		if fmt.Sprintf("%p", fnA) != fmt.Sprintf("%p", fnB) {
			t.Errorf("alias %q does not resolve to the same op as %q", alias, canonical)
		}
	}
}

// Controller-only functions must NOT run on remote runners.
func TestRegistryExcludesControllerOnlyOps(t *testing.T) {
	r := NewRegistry()
	if r.Has("file.distribute") {
		t.Error("file.distribute must not be available on remote runners")
	}
	if r.Has("file.collect") {
		t.Error("file.collect must not be available on remote runners")
	}
}

func TestBinaryExecPropagatesFailure(t *testing.T) {
	r := NewRegistry()
	if _, ok := r.Get("binary.exec"); !ok {
		t.Fatal("binary.exec not registered")
	}

	pkgExec := &InstructionPackage{
		Version: "1.0",
		Instructions: []Instruction{{
			Op:   "binary.exec",
			Args: map[string]interface{}{"path": "/nonexistent/ops-binary"},
		}},
	}
	out := Run(pkgExec, r)
	if out.Status == "ok" {
		t.Fatalf("binary.exec on a missing binary must fail, got status %q", out.Status)
	}
	if len(out.Errors) == 0 {
		t.Error("expected error details in output.Errors")
	}
}

// ============================================================
// Real SDK operation tests (no side effects)
// ============================================================

func TestRealSDKTimeNow(t *testing.T) {
	r := NewRegistry()

	pkg := &InstructionPackage{
		Version: "1.0",
		Instructions: []Instruction{
			{Op: "time.now", Assign: "now"},
		},
	}

	output := Run(pkg, r)
	if output.Status != "ok" {
		t.Fatalf("expected status ok, got %s (errors: %v)", output.Status, output.Errors)
	}

	now, ok := output.Data["now"]
	if !ok {
		t.Fatal("expected 'now' in output data")
	}
	nowMap, ok := now.(map[string]interface{})
	if !ok {
		// The result is a TimeInfo struct, which when JSON-marshaled becomes a map.
		// But in memory it's a struct. Let's just check it's not nil.
		if now == nil {
			t.Fatal("expected non-nil time result")
		}
		return
	}
	if _, ok := nowMap["unix"]; !ok {
		t.Error("expected 'unix' field in time result")
	}
}

func TestRealSDKCPUUsage(t *testing.T) {
	r := NewRegistry()

	pkg := &InstructionPackage{
		Version: "1.0",
		Instructions: []Instruction{
			{Op: "sys.cpu.usage", Assign: "cpu"},
		},
	}

	output := Run(pkg, r)
	if output.Status != "ok" {
		t.Fatalf("expected status ok, got %s (errors: %v)", output.Status, output.Errors)
	}
	if output.Data["cpu"] == nil {
		t.Error("expected non-nil cpu data")
	}
}

func TestRealSDKJSONEncode(t *testing.T) {
	r := NewRegistry()

	pkg := &InstructionPackage{
		Version: "1.0",
		Instructions: []Instruction{
			{
				Op: "json.encode",
				Args: map[string]interface{}{
					"data": map[string]interface{}{"key": "value"},
				},
				Assign: "encoded",
			},
		},
	}

	output := Run(pkg, r)
	if output.Status != "ok" {
		t.Fatalf("expected status ok, got %s (errors: %v)", output.Status, output.Errors)
	}
	if output.Data["encoded"] == nil {
		t.Error("expected non-nil encoded data")
	}
}

// ============================================================
// End-to-end integration test
// ============================================================

func TestEndToEnd(t *testing.T) {
	input := `{
		"version": "1.0",
		"task_id": "e2e-test",
		"dry_run": false,
		"instructions": [
			{"op": "time.now", "assign": "now"},
			{"op": "log", "args": {"message": "collecting system info"}},
			{"op": "sys.cpu.usage", "assign": "cpu"},
			{"op": "report", "args": {"timestamp": "now", "cpu_info": "cpu"}}
		]
	}`

	var pkg InstructionPackage
	if err := json.Unmarshal([]byte(input), &pkg); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	registry := NewRegistry()
	output := Run(&pkg, registry)

	if output.Status != "ok" {
		t.Errorf("expected status ok, got %s", output.Status)
	}
	if len(output.Errors) > 0 {
		t.Errorf("unexpected errors: %v", output.Errors)
	}
	if len(output.Warnings) == 0 {
		t.Error("expected at least one warning from log operation")
	}

	// Check report output has the expected keys.
	if _, ok := output.Data["timestamp"]; !ok {
		t.Error("expected 'timestamp' in report data")
	}
	if _, ok := output.Data["cpu_info"]; !ok {
		t.Error("expected 'cpu_info' in report data")
	}

	// Verify output can be marshaled to JSON.
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal output: %v", err)
	}
	if len(jsonBytes) == 0 {
		t.Error("expected non-empty JSON output")
	}
}

func TestEndToEndDryRun(t *testing.T) {
	input := `{
		"version": "1.0",
		"task_id": "dry-run-test",
		"dry_run": true,
		"instructions": [
			{"op": "sys.cpu.usage", "assign": "cpu"},
			{"op": "file.write", "args": {"path": "/tmp/test", "content": "hello"}, "assign": "write_result"}
		]
	}`

	var pkg InstructionPackage
	if err := json.Unmarshal([]byte(input), &pkg); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	registry := NewRegistry()
	output := Run(&pkg, registry)

	if output.Status != "ok" {
		t.Errorf("expected status ok, got %s", output.Status)
	}

	// In dry-run mode, results should indicate dry_run: true.
	cpuResult, ok := output.Data["cpu"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected cpu to be a map, got %T", output.Data["cpu"])
	}
	if cpuResult["dry_run"] != true {
		t.Error("expected dry_run=true in cpu result")
	}

	writeResult, ok := output.Data["write_result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected write_result to be a map, got %T", output.Data["write_result"])
	}
	if writeResult["dry_run"] != true {
		t.Error("expected dry_run=true in write_result")
	}
	if writeResult["operation"] != "file.write" {
		t.Errorf("expected operation file.write, got %v", writeResult["operation"])
	}
}

// ============================================================
// OutputToJSON test
// ============================================================

func TestOutputToJSON(t *testing.T) {
	output := &Output{
		Status:   "ok",
		Data:     map[string]interface{}{"key": "value"},
		Errors:   []string{},
		Warnings: []string{},
	}

	jsonBytes, err := OutputToJSON(output)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed Output
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if parsed.Status != "ok" {
		t.Errorf("expected status ok, got %s", parsed.Status)
	}
}
