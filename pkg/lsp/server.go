// Package lsp 实现 OpsLang 的 LSP 语言服务器
//
// 支持功能:
//   - 实时语法诊断（语法错误高亮）
//   - 关键字 / 内置函数 / 标准库模块补全
//   - 悬停信息（类型与文档）
//   - 跳转到函数定义
package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/opslang/opslang/pkg/lexer"
	"github.com/opslang/opslang/pkg/parser"
)

// --- JSON-RPC 传输层 ---

// RPCMessage JSON-RPC 消息
type RPCMessage struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *RPCError        `json:"error,omitempty"`
}

// RPCError JSON-RPC 错误
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Transport JSON-RPC 传输（基于 Content-Length 头）
type Transport struct {
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex
}

// NewTransport 创建传输层
func NewTransport(r io.Reader, w io.Writer) *Transport {
	return &Transport{
		reader: bufio.NewReader(r),
		writer: w,
	}
}

// ReadMessage 读取一条 JSON-RPC 消息
func (t *Transport) ReadMessage() (*RPCMessage, error) {
	// 读取头部
	var contentLength int
	for {
		line, err := t.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break // 头部结束
		}
		if strings.HasPrefix(line, "Content-Length:") {
			lenStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, err = strconv.Atoi(lenStr)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %s", lenStr)
			}
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	// 读取消息体
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(t.reader, body); err != nil {
		return nil, err
	}

	var msg RPCMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// SendMessage 发送一条 JSON-RPC 消息
func (t *Transport) SendMessage(msg *RPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	msg.JSONRPC = "2.0"
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := io.WriteString(t.writer, header); err != nil {
		return err
	}
	if _, err := t.writer.Write(body); err != nil {
		return err
	}
	return nil
}

// SendResponse 发送响应
func (t *Transport) SendResponse(id *json.RawMessage, result interface{}) error {
	return t.SendMessage(&RPCMessage{
		ID:     id,
		Result: result,
	})
}

// SendError 发送错误响应
func (t *Transport) SendError(id *json.RawMessage, code int, message string) error {
	return t.SendMessage(&RPCMessage{
		ID: id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	})
}

// SendNotification 发送通知（无 ID）
func (t *Transport) SendNotification(method string, params interface{}) error {
	paramsRaw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return t.SendMessage(&RPCMessage{
		Method: method,
		Params: paramsRaw,
	})
}

// --- LSP 类型定义 ---

// Position LSP 位置（0-based）
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range LSP 范围
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location LSP 位置
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Diagnostic LSP 诊断
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"` // 1=Error, 2=Warning, 3=Info, 4=Hint
	Message  string `json:"message"`
	Source   string `json:"source"`
}

// CompletionItem LSP 补全项
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"` // 1=Text, 2=Method, 3=Function, 6=Variable, 14=Keyword, 20=Constant
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

// Hover LSP 悬停信息
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent 富文本内容
type MarkupContent struct {
	Kind  string `json:"kind"` // "plaintext" or "markdown"
	Value string `json:"value"`
}

// --- 参数类型 ---

// InitializeParams 初始化参数
type InitializeParams struct {
	RootURI string `json:"rootUri"`
}

// InitializeResult 初始化结果
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

// ServerInfo 服务器信息
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerCapabilities 服务器能力
type ServerCapabilities struct {
	TextDocumentSync    int                      `json:"textDocumentSync"` // 1=Full
	CompletionProvider  *CompletionOptions       `json:"completionProvider,omitempty"`
	HoverProvider       bool                     `json:"hoverProvider"`
	DefinitionProvider  bool                     `json:"definitionProvider"`
	DiagnosticProvider  *DiagnosticServerOptions `json:"diagnosticProvider,omitempty"`
}

// CompletionOptions 补全选项
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters"`
	ResolveProvider   bool     `json:"resolveProvider"`
}

// DiagnosticServerOptions 诊断选项
type DiagnosticServerOptions struct {
	Identifier            string `json:"identifier"`
	InterFileDependencies bool   `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool   `json:"workspaceDiagnostics"`
}

// TextDocumentItem 文档项
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// DidOpenParams 打开文档参数
type DidOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidChangeParams 文档变更参数
type DidChangeParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// DidCloseParams 关闭文档参数
type DidCloseParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// TextDocumentIdentifier 文档标识
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// VersionedTextDocumentIdentifier 带版本的文档标识
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// TextDocumentContentChangeEvent 文档变更事件
type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

// CompletionParams 补全参数
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// CompletionList 补全列表
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// TextDocumentPositionParams 文档位置参数
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// PublishDiagnosticsParams 发布诊断参数
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// --- 服务器 ---

// Server LSP 服务器
type Server struct {
	transport *Transport
	documents map[string]string // URI -> 内容
	definitions map[string][]Definition // 函数定义索引
	logger    *os.File
}

// Definition 函数定义
type Definition struct {
	Name  string
	URI   string
	Range Range
}

// NewServer 创建服务器
func NewServer() *Server {
	return &Server{
		transport:   NewTransport(os.Stdin, os.Stdout),
		documents:   make(map[string]string),
		definitions: make(map[string][]Definition),
	}
}

// Run 运行服务器
func (s *Server) Run() error {
	for {
		msg, err := s.transport.ReadMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if err := s.handleMessage(msg); err != nil {
			return err
		}
	}
}

// handleMessage 处理消息
func (s *Server) handleMessage(msg *RPCMessage) error {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		return nil // 通知，无需响应
	case "shutdown":
		return s.transport.SendResponse(msg.ID, nil)
	case "exit":
		os.Exit(0)
		return nil
	case "textDocument/didOpen":
		return s.handleDidOpen(msg)
	case "textDocument/didChange":
		return s.handleDidChange(msg)
	case "textDocument/didClose":
		return s.handleDidClose(msg)
	case "textDocument/completion":
		return s.handleCompletion(msg)
	case "textDocument/hover":
		return s.handleHover(msg)
	case "textDocument/definition":
		return s.handleDefinition(msg)
	default:
		// 未实现的方法，返回空响应
		if msg.ID != nil {
			return s.transport.SendResponse(msg.ID, nil)
		}
		return nil
	}
}

// handleInitialize 处理初始化请求
func (s *Server) handleInitialize(msg *RPCMessage) error {
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: 1, // Full
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{".", "(", ","},
				ResolveProvider:   false,
			},
			HoverProvider:      true,
			DefinitionProvider: true,
		},
		ServerInfo: ServerInfo{
			Name:    "opslang-lsp",
			Version: "0.1.0",
		},
	}
	return s.transport.SendResponse(msg.ID, result)
}

// handleDidOpen 处理文档打开
func (s *Server) handleDidOpen(msg *RPCMessage) error {
	var params DidOpenParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return err
	}
	s.documents[params.TextDocument.URI] = params.TextDocument.Text
	s.indexDefinitions(params.TextDocument.URI, params.TextDocument.Text)
	return s.publishDiagnostics(params.TextDocument.URI, params.TextDocument.Text)
}

// handleDidChange 处理文档变更
func (s *Server) handleDidChange(msg *RPCMessage) error {
	var params DidChangeParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return err
	}
	if len(params.ContentChanges) == 0 {
		return nil
	}
	// Full sync: 取最后一个变更
	text := params.ContentChanges[len(params.ContentChanges)-1].Text
	s.documents[params.TextDocument.URI] = text
	s.indexDefinitions(params.TextDocument.URI, text)
	return s.publishDiagnostics(params.TextDocument.URI, text)
}

// handleDidClose 处理文档关闭
func (s *Server) handleDidClose(msg *RPCMessage) error {
	var params DidCloseParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return err
	}
	delete(s.documents, params.TextDocument.URI)
	delete(s.definitions, params.TextDocument.URI)
	return nil
}

// publishDiagnostics 发布诊断
func (s *Server) publishDiagnostics(uri, text string) error {
	diagnostics := s.analyze(text)
	return s.transport.SendNotification("textDocument/publishDiagnostics",
		PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diagnostics,
		})
}

// analyze 分析代码，返回诊断
func (s *Server) analyze(text string) []Diagnostic {
	var diagnostics []Diagnostic

	// 词法分析
	l := lexer.New(text, "document")
	tokens := l.Tokenize()

	// 语法分析
	p := parser.New(tokens)
	_, err := p.Parse()
	if err != nil {
		// 将解析错误转为诊断
		if pe, ok := err.(*parser.ParseError); ok {
			line := pe.Line - 1
			if line < 0 {
				line = 0
			}
			col := pe.Column - 1
			if col < 0 {
				col = 0
			}
			diagnostics = append(diagnostics, Diagnostic{
				Range: Range{
					Start: Position{Line: line, Character: col},
					End:   Position{Line: line, Character: col + 1},
				},
				Severity: 1, // Error
				Message:  pe.Message,
				Source:   "opslang",
			})
		}
	}

	return diagnostics
}

// indexDefinitions 索引函数定义
func (s *Server) indexDefinitions(uri, text string) {
	var defs []Definition
	lines := strings.Split(text, "\n")
	for lineNo, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "fn ") {
			// 提取函数名
			rest := strings.TrimPrefix(trimmed, "fn ")
			parts := strings.SplitN(rest, "(", 2)
			if len(parts) >= 1 {
				name := strings.TrimSpace(parts[0])
				if name != "" {
					col := strings.Index(line, name)
					defs = append(defs, Definition{
						Name: name,
						URI:  uri,
						Range: Range{
							Start: Position{Line: lineNo, Character: col},
							End:   Position{Line: lineNo, Character: col + len(name)},
						},
					})
				}
			}
		}
	}
	s.definitions[uri] = defs
}

// handleCompletion 处理补全请求
func (s *Server) handleCompletion(msg *RPCMessage) error {
	var params CompletionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return err
	}

	items := s.getCompletions(params.TextDocument.URI, params.Position)
	return s.transport.SendResponse(msg.ID, CompletionList{
		IsIncomplete: false,
		Items:        items,
	})
}

// getCompletions 获取补全项
func (s *Server) getCompletions(uri string, pos Position) []CompletionItem {
	var items []CompletionItem

	// 关键字
	keywords := []struct {
		label, detail string
	}{
		{"fn", "函数定义"},
		{"return", "返回语句"},
		{"if", "条件语句"},
		{"else", "else 分支"},
		{"for", "for 循环"},
		{"in", "迭代关键字"},
		{"while", "while 循环"},
		{"break", "跳出循环"},
		{"continue", "继续下一次循环"},
		{"try", "try 块"},
		{"catch", "catch 块"},
		{"import", "导入模块"},
		{"true", "布尔真"},
		{"false", "布尔假"},
		{"nil", "空值"},
	}
	for _, kw := range keywords {
		items = append(items, CompletionItem{
			Label:  kw.label,
			Kind:   14, // Keyword
			Detail: kw.detail,
		})
	}

	// 内置函数
	builtins := []struct {
		label, detail, doc string
	}{
		{"print", "print(args...)", "输出到标准输出"},
		{"len", "len(obj)", "获取长度（字符串/数组/字典）"},
		{"type", "type(obj)", "获取类型名"},
		{"str", "str(obj)", "转换为字符串"},
		{"int", "int(obj)", "转换为整数"},
		{"range", "range([start], stop, [step])", "生成整数序列"},
		{"append", "append(arr, elem)", "追加元素到数组"},
		{"map", "map(arr, fn)", "对数组每个元素应用函数"},
		{"filter", "filter(arr, fn)", "过滤数组元素"},
	}
	for _, b := range builtins {
		items = append(items, CompletionItem{
			Label:         b.label,
			Kind:          3, // Function
			Detail:        b.detail,
			Documentation: b.doc,
		})
	}

	// 标准库模块
	modules := []struct {
		label, doc string
	}{
		{"file", "文件操作 (read/write/exists/mkdir/list)"},
		{"process", "进程与环境 (shell/run/env/cwd/hostname)"},
		{"ssh", "远程执行 (run/copy/ping)"},
		{"fleet", "批量执行 (parallel/serial/exec/summary)"},
		{"json", "JSON 处理 (parse/dump/prettify)"},
		{"yaml", "YAML 处理 (parse/dump/load_file/save_file)"},
		{"toml", "TOML 处理 (parse/load_file)"},
		{"strings", "字符串工具 (split/join/contains/replace/trim/upper/lower)"},
		{"math", "数学函数 (abs/min/max)"},
		{"ensure", "声明式管理 (file/dir/line/service/package/user)"},
		{"inventory", "主机清单 (load/group/all/from_list)"},
	}
	for _, m := range modules {
		items = append(items, CompletionItem{
			Label:         m.label,
			Kind:          9, // Module
			Detail:        "标准库模块",
			Documentation: m.doc,
		})
	}

	// 当前文档中的函数定义
	if defs, ok := s.definitions[uri]; ok {
		for _, d := range defs {
			items = append(items, CompletionItem{
				Label: d.Name,
				Kind:  3, // Function
				Detail: "用户定义函数",
			})
		}
	}

	return items
}

// handleHover 处理悬停请求
func (s *Server) handleHover(msg *RPCMessage) error {
	var params TextDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return err
	}

	text, ok := s.documents[params.TextDocument.URI]
	if !ok {
		return s.transport.SendResponse(msg.ID, nil)
	}

	word := s.getWordAt(text, params.Position)
	if word == "" {
		return s.transport.SendResponse(msg.ID, nil)
	}

	hover := s.getHoverInfo(word)
	if hover == nil {
		return s.transport.SendResponse(msg.ID, nil)
	}

	return s.transport.SendResponse(msg.ID, hover)
}

// getWordAt 获取指定位置的单词
func (s *Server) getWordAt(text string, pos Position) string {
	lines := strings.Split(text, "\n")
	if pos.Line >= len(lines) {
		return ""
	}
	line := lines[pos.Line]
	if pos.Character >= len(line) {
		return ""
	}

	// 向前找起点
	start := pos.Character
	for start > 0 {
		ch := line[start-1]
		if !isIdentChar(ch) {
			break
		}
		start--
	}

	// 向后找终点
	end := pos.Character
	for end < len(line) {
		ch := line[end]
		if !isIdentChar(ch) {
			break
		}
		end++
	}

	if start == end {
		return ""
	}
	return line[start:end]
}

func isIdentChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '_'
}

// getHoverInfo 获取悬停信息
func (s *Server) getHoverInfo(word string) *Hover {
	// 内置函数文档
	builtinDocs := map[string]string{
		"print":    "```ops\nprint(args...)\n```\n输出参数到标准输出，多个参数以空格分隔。",
		"len":      "```ops\nlen(obj) -> int\n```\n获取字符串、数组或字典的长度。",
		"type":     "```ops\ntype(obj) -> string\n```\n返回值的类型名（nil/bool/int/float/string/array/map/function）。",
		"str":      "```ops\nstr(obj) -> string\n```\n将任意值转换为字符串表示。",
		"int":      "```ops\nint(obj) -> int\n```\n将字符串或浮点数转换为整数。",
		"range":    "```ops\nrange(stop) -> array\nrange(start, stop) -> array\nrange(start, stop, step) -> array\n```\n生成整数序列。",
		"append":   "```ops\nappend(arr, elem) -> array\n```\n返回追加了元素的新数组（原数组不变）。",
		"map":      "```ops\nmap(arr, fn) -> array\n```\n对数组每个元素应用函数，返回新数组。",
		"filter":   "```ops\nfilter(arr, fn) -> array\n```\n过滤数组，只保留函数返回 true 的元素。",
	}

	if doc, ok := builtinDocs[word]; ok {
		return &Hover{
			Contents: MarkupContent{Kind: "markdown", Value: doc},
		}
	}

	// 模块文档
	moduleDocs := map[string]string{
		"file":      "**file** — 文件操作\n\n- `file.read(path)` 读取文件\n- `file.write(path, content)` 写入文件\n- `file.exists(path)` 检查存在\n- `file.mkdir(path)` 创建目录\n- `file.list(dir)` 列出目录\n- `file.basename(path)` 文件名\n- `file.dirname(path)` 目录名",
		"process":   "**process** — 进程与环境\n\n- `process.shell(cmd)` 执行 shell 命令\n- `process.run(cmd, args...)` 直接执行\n- `process.env(name)` 读取环境变量\n- `process.cwd()` 当前目录\n- `process.hostname()` 主机名",
		"ssh":       "**ssh** — 远程执行\n\n- `ssh.run(host, cmd, [user])` 远程执行\n- `ssh.copy(local, host, remote)` SCP 传输\n- `ssh.ping(host)` 连通性检测",
		"fleet":     "**fleet** — 批量执行引擎\n\n- `fleet.parallel(hosts, fn, [n])` 并发执行\n- `fleet.serial(hosts, fn)` 串行执行\n- `fleet.exec(hosts, cmd, [user])` 批量 SSH\n- `fleet.summary(results)` 结果汇总",
		"json":      "**json** — JSON 处理\n\n- `json.parse(str)` 解析\n- `json.dump(value)` 序列化\n- `json.prettify(value)` 美化输出\n- `json.load_file(path)` 读文件\n- `json.save_file(path, value)` 写文件",
		"yaml":      "**yaml** — YAML 处理\n\n- `yaml.parse(str)` 解析\n- `yaml.dump(value)` 序列化\n- `yaml.load_file(path)` 读文件\n- `yaml.save_file(path, value)` 写文件",
		"toml":      "**toml** — TOML 处理\n\n- `toml.parse(str)` 解析\n- `toml.load_file(path)` 读文件",
		"strings":   "**strings** — 字符串工具\n\n- `strings.split(s, sep)` 分割\n- `strings.join(arr, sep)` 连接\n- `strings.contains(s, sub)` 包含检测\n- `strings.replace(s, old, new)` 替换\n- `strings.trim(s)` 去空白\n- `strings.upper(s)` 转大写\n- `strings.lower(s)` 转小写\n- `strings.has_prefix(s, prefix)` 前缀检测\n- `strings.has_suffix(s, suffix)` 后缀检测",
		"math":      "**math** — 数学函数\n\n- `math.abs(n)` 绝对值\n- `math.min(a, b)` 最小值\n- `math.max(a, b)` 最大值",
		"ensure":    "**ensure** — 声明式资源管理（幂等）\n\n- `ensure.file(path, content, [mode])` 文件\n- `ensure.dir(path, [mode])` 目录\n- `ensure.line(path, line)` 行内容\n- `ensure.service(name, state, enabled)` 服务\n- `ensure.package(name, [state])` 包\n- `ensure.user(name, [shell], [groups])` 用户",
		"inventory": "**inventory** — 主机清单\n\n- `inventory.load(path)` 从 INI 文件加载\n- `inventory.from_list(hosts)` 从数组创建\n- `inventory.group(inv, name)` 获取分组\n- `inventory.all(inv)` 所有主机",
	}

	if doc, ok := moduleDocs[word]; ok {
		return &Hover{
			Contents: MarkupContent{Kind: "markdown", Value: doc},
		}
	}

	// 关键字
	keywordDocs := map[string]string{
		"fn":       "**fn** — 函数定义\n\n```ops\nfn name(args)\n    body\n```",
		"if":       "**if** — 条件语句\n\n```ops\nif condition\n    body\nelse if condition2\n    body2\nelse\n    body3\n```",
		"for":      "**for** — 循环\n\n```ops\nfor item in collection\n    body\n```",
		"while":    "**while** — 条件循环\n\n```ops\nwhile condition\n    body\n```",
		"try":      "**try** — 错误处理\n\n```ops\ntry\n    risky_code()\ncatch e\n    handle_error(e)\n```",
		"return":   "**return** — 从函数返回值\n\n```ops\nfn add(a, b)\n    return a + b\n```",
		"break":    "**break** — 跳出当前循环",
		"continue": "**continue** — 继续下一次循环迭代",
		"import":   "**import** — 导入模块\n\n```ops\nimport module_name\nimport module_name as alias\n```",
	}

	if doc, ok := keywordDocs[word]; ok {
		return &Hover{
			Contents: MarkupContent{Kind: "markdown", Value: doc},
		}
	}

	return nil
}

// handleDefinition 处理跳转到定义
func (s *Server) handleDefinition(msg *RPCMessage) error {
	var params TextDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return err
	}

	text, ok := s.documents[params.TextDocument.URI]
	if !ok {
		return s.transport.SendResponse(msg.ID, nil)
	}

	word := s.getWordAt(text, params.Position)
	if word == "" {
		return s.transport.SendResponse(msg.ID, nil)
	}

	// 在当前文档中查找函数定义
	if defs, ok := s.definitions[params.TextDocument.URI]; ok {
		for _, d := range defs {
			if d.Name == word {
				loc := Location{URI: d.URI, Range: d.Range}
				return s.transport.SendResponse(msg.ID, loc)
			}
		}
	}

	return s.transport.SendResponse(msg.ID, nil)
}
