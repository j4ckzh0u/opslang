package interpreter

import (
	"strings"
	"testing"
)

func TestStructToMap(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantNil bool
		wantErr bool
	}{
		{
			name:    "nil input",
			input:   nil,
			wantNil: true,
		},
		{
			name:  "simple struct",
			input: struct{ Name string }{Name: "test"},
		},
		{
			name: "struct with multiple fields",
			input: struct {
				A int
				B string
				C bool
			}{A: 1, B: "hello", C: true},
		},
		{
			name: "nested struct",
			input: struct {
				Inner struct{ Value int }
			}{Inner: struct{ Value int }{Value: 42}},
		},
		{
			name:  "map input",
			input: map[string]int{"a": 1},
		},
		{
			name:    "unmarshallable input",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := structToMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("structToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil && got != nil {
				t.Errorf("structToMap() = %v, want nil", got)
			}
			if !tt.wantNil && !tt.wantErr && got == nil {
				t.Errorf("structToMap() = nil, want non-nil")
			}
		})
	}
}

func TestStructToMapValues(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	result, err := structToMap(testStruct{Name: "cpu", Value: 42})
	if err != nil {
		t.Fatalf("structToMap() error = %v", err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("structToMap() result is not a map, got %T", result)
	}
	if m["name"] != "cpu" {
		t.Errorf("name = %v, want 'cpu'", m["name"])
	}
	// JSON unmarshaling converts numbers to float64
	if m["value"] != float64(42) {
		t.Errorf("value = %v, want 42", m["value"])
	}
}

func TestRegisterSDKBuiltins(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	// Verify all expected builtins are registered
	expectedBuiltins := []string{
		// sys
		"sys.hostname", "sys.cpu.usage", "sys.cpu.info", "sys.cpu.count",
		"sys.memory.info", "sys.disk.usage", "sys.disk.partitions",
		"sys.load", "sys.os", "sys.uptime", "sys.users", "sys.net.interfaces",
		// file
		"file.read", "file.write", "file.exists", "file.copy", "file.move",
		"file.delete", "file.stat", "file.list", "file.mkdir", "file.checksum",
		"file.distribute", "file.collect",
		// net
		"net.http_get", "net.http_post", "net.tcp_check", "net.dns_lookup", "net.interfaces",
		// process
		"process.list", "process.find_by_name", "process.find_by_port", "process.exec",
		// service
		"service.status", "service.start", "service.stop", "service.restart", "service.enable",
		// time
		"time.now", "time.format", "time.since", "time.sleep",
		// json
		"json.encode", "json.decode",
		// yaml
		"yaml.encode", "yaml.decode",
	}

	for _, name := range expectedBuiltins {
		if _, ok := interp.builtins[name]; !ok {
			t.Errorf("builtin %q not registered", name)
		}
	}
}

func TestSDKBuiltinArgValidation(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	// Test argument count/type validation for builtins that check args
	tests := []struct {
		name    string
		builtin string
		args    []interface{}
		wantErr string
	}{
		// sys.disk.usage
		{"sys.disk.usage no args", "sys.disk.usage", nil, "requires 1 argument"},
		{"sys.disk.usage wrong type", "sys.disk.usage", []interface{}{123}, "argument must be string"},

		// file.read
		{"file.read no args", "file.read", nil, "requires 1 argument"},
		{"file.read wrong type", "file.read", []interface{}{123}, "argument must be string"},

		// file.write
		{"file.write no args", "file.write", nil, "requires 2 arguments"},
		{"file.write path wrong type", "file.write", []interface{}{123, "content"}, "path must be string"},
		{"file.write content wrong type", "file.write", []interface{}{"/tmp/test", 123}, "content must be string"},

		// file.exists
		{"file.exists no args", "file.exists", nil, "requires 1 argument"},
		{"file.exists wrong type", "file.exists", []interface{}{123}, "argument must be string"},

		// file.copy
		{"file.copy no args", "file.copy", nil, "requires 2 arguments"},
		{"file.copy src wrong type", "file.copy", []interface{}{123, "/dst"}, "src must be string"},
		{"file.copy dst wrong type", "file.copy", []interface{}{"/src", 123}, "dst must be string"},

		// file.move
		{"file.move no args", "file.move", nil, "requires 2 arguments"},
		{"file.move src wrong type", "file.move", []interface{}{123, "/dst"}, "src must be string"},
		{"file.move dst wrong type", "file.move", []interface{}{"/src", 123}, "dst must be string"},

		// file.delete
		{"file.delete no args", "file.delete", nil, "requires 1 argument"},
		{"file.delete wrong type", "file.delete", []interface{}{123}, "argument must be string"},

		// file.stat
		{"file.stat no args", "file.stat", nil, "requires 1 argument"},
		{"file.stat wrong type", "file.stat", []interface{}{123}, "argument must be string"},

		// file.list
		{"file.list no args", "file.list", nil, "requires 1 argument"},
		{"file.list wrong type", "file.list", []interface{}{123}, "argument must be string"},

		// file.mkdir
		{"file.mkdir no args", "file.mkdir", nil, "requires 1 argument"},
		{"file.mkdir wrong type", "file.mkdir", []interface{}{123}, "argument must be string"},

		// file.checksum
		{"file.checksum no args", "file.checksum", nil, "requires 2 arguments"},
		{"file.checksum path wrong type", "file.checksum", []interface{}{123, "sha256"}, "path must be string"},
		{"file.checksum algo wrong type", "file.checksum", []interface{}{"/tmp/test", 123}, "algo must be string"},

		// file.distribute
		{"file.distribute no args", "file.distribute", nil, "requires at least 2 arguments"},
		{"file.distribute source wrong type", "file.distribute", []interface{}{123, []interface{}{}}, "source must be string"},
		{"file.distribute targets wrong type", "file.distribute", []interface{}{"/src", "not-a-list"}, "targets must be a list"},

		// file.collect
		{"file.collect no args", "file.collect", nil, "requires at least 2 arguments"},
		{"file.collect source wrong type", "file.collect", []interface{}{123, []interface{}{}}, "source must be string"},
		{"file.collect targets wrong type", "file.collect", []interface{}{"/src", "not-a-list"}, "targets must be a list"},

		// net.http_get
		{"net.http_get no args", "net.http_get", nil, "requires 1 argument"},
		{"net.http_get wrong type", "net.http_get", []interface{}{123}, "argument must be string"},

		// net.http_post
		{"net.http_post no args", "net.http_post", nil, "requires 2 arguments"},
		{"net.http_post url wrong type", "net.http_post", []interface{}{123, "body"}, "url must be string"},
		{"net.http_post body wrong type", "net.http_post", []interface{}{"http://example.com", 123}, "body must be string"},

		// net.tcp_check
		{"net.tcp_check no args", "net.tcp_check", nil, "requires 2 arguments"},
		{"net.tcp_check host wrong type", "net.tcp_check", []interface{}{123, 80}, "host must be string"},
		{"net.tcp_check port wrong type", "net.tcp_check", []interface{}{"localhost", "not-a-number"}, "port must be number"},

		// net.dns_lookup
		{"net.dns_lookup no args", "net.dns_lookup", nil, "requires 1 argument"},
		{"net.dns_lookup wrong type", "net.dns_lookup", []interface{}{123}, "argument must be string"},

		// process.find_by_name
		{"process.find_by_name no args", "process.find_by_name", nil, "requires 1 argument"},
		{"process.find_by_name wrong type", "process.find_by_name", []interface{}{123}, "argument must be string"},

		// process.find_by_port
		{"process.find_by_port no args", "process.find_by_port", nil, "requires 1 argument"},
		{"process.find_by_port wrong type", "process.find_by_port", []interface{}{"not-a-number"}, "argument must be number"},

		// process.exec
		{"process.exec no args", "process.exec", nil, "requires at least 1 argument"},
		{"process.exec wrong type", "process.exec", []interface{}{123}, "command must be string"},

		// service.*
		{"service.status no args", "service.status", nil, "requires 1 argument"},
		{"service.status wrong type", "service.status", []interface{}{123}, "argument must be string"},
		{"service.start no args", "service.start", nil, "requires 1 argument"},
		{"service.start wrong type", "service.start", []interface{}{123}, "argument must be string"},
		{"service.stop no args", "service.stop", nil, "requires 1 argument"},
		{"service.stop wrong type", "service.stop", []interface{}{123}, "argument must be string"},
		{"service.restart no args", "service.restart", nil, "requires 1 argument"},
		{"service.restart wrong type", "service.restart", []interface{}{123}, "argument must be string"},
		{"service.enable no args", "service.enable", nil, "requires 1 argument"},
		{"service.enable wrong type", "service.enable", []interface{}{123}, "argument must be string"},

		// time.format
		{"time.format no args", "time.format", nil, "requires 2 arguments"},
		{"time.format unix wrong type", "time.format", []interface{}{"not-a-number", "2006-01-02"}, "unix must be number"},
		{"time.format layout wrong type", "time.format", []interface{}{float64(1000000), 123}, "layout must be string"},

		// time.since
		{"time.since no args", "time.since", nil, "requires 1 argument"},
		{"time.since wrong type", "time.since", []interface{}{"not-a-number"}, "argument must be number"},

		// time.sleep
		{"time.sleep no args", "time.sleep", nil, "requires 1 argument"},
		{"time.sleep wrong type", "time.sleep", []interface{}{"not-a-number"}, "argument must be number"},

		// json
		{"json.encode no args", "json.encode", nil, "requires 1 argument"},
		{"json.decode no args", "json.decode", nil, "requires 1 argument"},
		{"json.decode wrong type", "json.decode", []interface{}{123}, "argument must be string"},

		// yaml
		{"yaml.encode no args", "yaml.encode", nil, "requires 1 argument"},
		{"yaml.decode no args", "yaml.decode", nil, "requires 1 argument"},
		{"yaml.decode wrong type", "yaml.decode", []interface{}{123}, "argument must be string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := interp.builtins[tt.builtin]
			if !ok {
				t.Fatalf("builtin %q not found", tt.builtin)
			}
			_, err := fn(tt.args...)
			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErr)
				return
			}
			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestSDKBuiltinJsonOperations(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	// json.encode
	fn := interp.builtins["json.encode"]
	result, err := fn(map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("json.encode() error = %v", err)
	}
	if result == nil {
		t.Error("json.encode() returned nil")
	}

	// json.decode
	fn = interp.builtins["json.decode"]
	result, err = fn(`{"key":"value"}`)
	if err != nil {
		t.Fatalf("json.decode() error = %v", err)
	}
	if result == nil {
		t.Error("json.decode() returned nil")
	}
}

func TestSDKBuiltinYamlOperations(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	// yaml.encode
	fn := interp.builtins["yaml.encode"]
	result, err := fn(map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("yaml.encode() error = %v", err)
	}
	if result == nil {
		t.Error("yaml.encode() returned nil")
	}

	// yaml.decode
	fn = interp.builtins["yaml.decode"]
	result, err = fn("key: value\n")
	if err != nil {
		t.Fatalf("yaml.decode() error = %v", err)
	}
	if result == nil {
		t.Error("yaml.decode() returned nil")
	}
}

func TestSDKBuiltinTimeNow(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["time.now"]
	result, err := fn()
	if err != nil {
		t.Fatalf("time.now() error = %v", err)
	}
	if result == nil {
		t.Error("time.now() returned nil")
	}
}

func TestSDKBuiltinTimeSince(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["time.since"]
	result, err := fn(float64(1000000))
	if err != nil {
		t.Fatalf("time.since() error = %v", err)
	}
	if result == nil {
		t.Error("time.since() returned nil")
	}
}

func TestSDKBuiltinTimeSleep(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["time.sleep"]
	result, err := fn(float64(1)) // 1ms
	if err != nil {
		t.Fatalf("time.sleep() error = %v", err)
	}
	if result == nil {
		t.Error("time.sleep() returned nil")
	}
}

func TestSDKBuiltinTimeFormat(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["time.format"]
	result, err := fn(float64(1700000000), "2006-01-02")
	if err != nil {
		t.Fatalf("time.format() error = %v", err)
	}
	if result == nil {
		t.Error("time.format() returned nil")
	}
}

func TestSDKBuiltinProcessList(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["process.list"]
	result, err := fn()
	if err != nil {
		t.Fatalf("process.list() error = %v", err)
	}
	if result == nil {
		t.Error("process.list() returned nil")
	}
}

func TestSDKBuiltinNetInterfaces(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["net.interfaces"]
	result, err := fn()
	if err != nil {
		t.Fatalf("net.interfaces() error = %v", err)
	}
	if result == nil {
		t.Error("net.interfaces() returned nil")
	}
}

func TestSDKBuiltinSysFunctions(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	// Test functions that don't require arguments and should work on any system
	noArgFuncs := []string{
		"sys.hostname", "sys.cpu.usage", "sys.cpu.info", "sys.cpu.count",
		"sys.memory.info", "sys.disk.partitions", "sys.load", "sys.os",
		"sys.uptime", "sys.users", "sys.net.interfaces",
	}

	for _, name := range noArgFuncs {
		t.Run(name, func(t *testing.T) {
			fn, ok := interp.builtins[name]
			if !ok {
				t.Fatalf("builtin %q not found", name)
			}
			result, err := fn()
			if err != nil {
				t.Fatalf("%s() error = %v", name, err)
			}
			if result == nil {
				t.Errorf("%s() returned nil", name)
			}
		})
	}
}

func TestSDKBuiltinSysDiskUsage(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["sys.disk.usage"]
	result, err := fn("/")
	if err != nil {
		t.Fatalf("sys.disk.usage() error = %v", err)
	}
	if result == nil {
		t.Error("sys.disk.usage() returned nil")
	}
}

func TestSDKBuiltinFileExists(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["file.exists"]
	result, err := fn("/tmp")
	if err != nil {
		t.Fatalf("file.exists() error = %v", err)
	}
	if result == nil {
		t.Error("file.exists() returned nil")
	}
}

func TestSDKBuiltinFileStat(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["file.stat"]
	result, err := fn("/tmp")
	if err != nil {
		t.Fatalf("file.stat() error = %v", err)
	}
	if result == nil {
		t.Error("file.stat() returned nil")
	}
}

func TestSDKBuiltinFileList(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["file.list"]
	result, err := fn("/tmp")
	if err != nil {
		t.Fatalf("file.list() error = %v", err)
	}
	if result == nil {
		t.Error("file.list() returned nil")
	}
}

func TestSDKBuiltinFileDistributeTargetValidation(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["file.distribute"]

	// Test with invalid target item (not a dict)
	_, err := fn("/src", []interface{}{"not-a-dict"})
	if err == nil {
		t.Error("expected error for non-dict target item")
	}
}

func TestSDKBuiltinFileCollectTargetValidation(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["file.collect"]

	// Test with invalid target item (not a dict)
	_, err := fn("/src", []interface{}{"not-a-dict"})
	if err == nil {
		t.Error("expected error for non-dict target item")
	}
}

func TestSDKBuiltinFileTransferResumeOptionValidation(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)
	tests := []struct {
		name      string
		builtin   string
		options   interface{}
		wantError string
	}{
		{name: "distribute options type", builtin: "file.distribute", options: "bad", wantError: "options must be a dict"},
		{name: "distribute resume type", builtin: "file.distribute", options: map[string]interface{}{"resume": "yes"}, wantError: "resume must be bool"},
		{name: "distribute retention type", builtin: "file.distribute", options: map[string]interface{}{"part_retention": "hour"}, wantError: "part_retention must be a number"},
		{name: "distribute negative retention", builtin: "file.distribute", options: map[string]interface{}{"part_retention": float64(-1)}, wantError: "part_retention must be a non-negative integer"},
		{name: "distribute relay type", builtin: "file.distribute", options: map[string]interface{}{"relay": "yes"}, wantError: "relay must be bool"},
		{name: "distribute relay threshold", builtin: "file.distribute", options: map[string]interface{}{"relay_threshold": float64(0)}, wantError: "relay_threshold must be a positive integer"},
		{name: "distribute target tags type", builtin: "file.distribute", options: map[string]interface{}{}, wantError: "tags must be a dict"},
		{name: "collect options type", builtin: "file.collect", options: "bad", wantError: "options must be a dict"},
		{name: "collect resume type", builtin: "file.collect", options: map[string]interface{}{"resume": float64(1)}, wantError: "resume must be bool"},
		{name: "collect fractional retention", builtin: "file.collect", options: map[string]interface{}{"part_retention": 1.5}, wantError: "part_retention must be a non-negative integer"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			targets := []interface{}{}
			if test.name == "distribute target tags type" {
				targets = []interface{}{map[string]interface{}{"host": "host1", "tags": "bad"}}
			}
			_, err := interp.builtins[test.builtin]("/source", targets, test.options)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("error = %v, want substring %q", err, test.wantError)
			}
		})
	}
}

func TestSDKBuiltinProcessExec(t *testing.T) {
	interp := New(nil)
	RegisterSDKBuiltins(interp)

	fn := interp.builtins["process.exec"]
	result, err := fn("echo", "hello")
	if err != nil {
		t.Fatalf("process.exec() error = %v", err)
	}
	if result == nil {
		t.Error("process.exec() returned nil")
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
