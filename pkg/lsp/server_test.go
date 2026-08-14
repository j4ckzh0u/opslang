package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// 构造 LSP 消息
func makeMessage(method string, id interface{}, params interface{}) string {
	msg := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if id != nil {
		msg["id"] = id
	}
	if params != nil {
		msg["params"] = params
	}
	body, _ := json.Marshal(msg)
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
}

func TestLSPInitialize(t *testing.T) {
	input := makeMessage("initialize", 1, map[string]interface{}{
		"rootUri": "file:///tmp/test",
	})
	input += makeMessage("initialized", nil, map[string]interface{}{})
	input += makeMessage("shutdown", 2, nil)

	reader := strings.NewReader(input)
	var output bytes.Buffer

	transport := NewTransport(reader, &output)
	server := &Server{
		transport:   transport,
		documents:   make(map[string]string),
		definitions: make(map[string][]Definition),
	}

	// 处理 initialize
	msg, err := server.transport.ReadMessage()
	if err != nil {
		t.Fatalf("读取消息失败: %v", err)
	}
	if msg.Method != "initialize" {
		t.Fatalf("期望 initialize, 实际 %s", msg.Method)
	}
	if err := server.handleMessage(msg); err != nil {
		t.Fatalf("处理失败: %v", err)
	}

	// 检查输出包含 ServerInfo
	outputStr := output.String()
	if !strings.Contains(outputStr, "opslang-lsp") {
		t.Errorf("响应中应包含服务器名 opslang-lsp:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "completionProvider") {
		t.Errorf("响应中应包含补全能力:\n%s", outputStr)
	}
}

func TestLSPDiagnostics(t *testing.T) {
	server := &Server{
		transport:   nil,
		documents:   make(map[string]string),
		definitions: make(map[string][]Definition),
	}

	// 有效代码，无错误
	diags := server.analyze(`x = 42
print(x)`)
	if len(diags) != 0 {
		t.Errorf("有效代码应有 0 个诊断, 实际 %d: %+v", len(diags), diags)
	}

	// 语法错误
	diags = server.analyze(`if
`)
	if len(diags) == 0 {
		t.Error("无效代码应有诊断")
	}
}

func TestLSPCompletions(t *testing.T) {
	server := &Server{
		transport:   nil,
		documents:   make(map[string]string),
		definitions: make(map[string][]Definition),
	}

	items := server.getCompletions("file:///test", Position{Line: 0, Character: 0})
	if len(items) == 0 {
		t.Fatal("应返回补全项")
	}

	// 检查包含关键字
	foundKeyword := false
	foundBuiltin := false
	foundModule := false
	for _, item := range items {
		if item.Label == "fn" && item.Kind == 14 {
			foundKeyword = true
		}
		if item.Label == "print" && item.Kind == 3 {
			foundBuiltin = true
		}
		if item.Label == "fleet" && item.Kind == 9 {
			foundModule = true
		}
	}

	if !foundKeyword {
		t.Error("应包含关键字补全")
	}
	if !foundBuiltin {
		t.Error("应包含内置函数补全")
	}
	if !foundModule {
		t.Error("应包含模块补全")
	}
}

func TestLSPHover(t *testing.T) {
	server := &Server{
		transport:   nil,
		documents:   make(map[string]string),
		definitions: make(map[string][]Definition),
	}

	// 内置函数 hover
	hover := server.getHoverInfo("print")
	if hover == nil {
		t.Fatal("print 应有 hover 信息")
	}
	if !strings.Contains(hover.Contents.Value, "print") {
		t.Error("hover 内容应包含 print")
	}

	// 模块 hover
	hover = server.getHoverInfo("fleet")
	if hover == nil {
		t.Fatal("fleet 应有 hover 信息")
	}
	if !strings.Contains(hover.Contents.Value, "parallel") {
		t.Error("fleet hover 应包含 parallel")
	}

	// 未知标识符
	hover = server.getHoverInfo("unknown_xyz")
	if hover != nil {
		t.Error("未知标识符不应有 hover 信息")
	}
}

func TestLSPDefinitionIndex(t *testing.T) {
	server := &Server{
		transport:   nil,
		documents:   make(map[string]string),
		definitions: make(map[string][]Definition),
	}

	code := `fn hello(name)
    print(name)

fn world()
    hello("World")`

	server.indexDefinitions("file:///test.ops", code)

	defs := server.definitions["file:///test.ops"]
	if len(defs) != 2 {
		t.Fatalf("应索引 2 个函数定义, 实际 %d", len(defs))
	}

	foundHello := false
	foundWorld := false
	for _, d := range defs {
		if d.Name == "hello" {
			foundHello = true
			if d.Range.Start.Line != 0 {
				t.Errorf("hello 应在第 0 行, 实际 %d", d.Range.Start.Line)
			}
		}
		if d.Name == "world" {
			foundWorld = true
			if d.Range.Start.Line != 3 {
				t.Errorf("world 应在第 3 行, 实际 %d", d.Range.Start.Line)
			}
		}
	}

	if !foundHello {
		t.Error("未找到 hello 函数定义")
	}
	if !foundWorld {
		t.Error("未找到 world 函数定义")
	}
}

func TestLSPGetWordAt(t *testing.T) {
	server := &Server{}

	tests := []struct {
		text     string
		pos      Position
		expected string
	}{
		{"hello world", Position{Line: 0, Character: 2}, "hello"},
		{"hello world", Position{Line: 0, Character: 7}, "world"},
		{"print(x)", Position{Line: 0, Character: 0}, "print"},
		{"fleet.parallel(hosts)", Position{Line: 0, Character: 8}, "parallel"},
		{"", Position{Line: 0, Character: 0}, ""},
	}

	for _, tt := range tests {
		result := server.getWordAt(tt.text, tt.pos)
		if result != tt.expected {
			t.Errorf("getWordAt(%q, %+v) = %q, 期望 %q", tt.text, tt.pos, result, tt.expected)
		}
	}
}
