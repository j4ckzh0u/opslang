package lexer

import (
	"testing"
)

func TestBasicTokens(t *testing.T) {
	input := `x = 42`
	l := New(input, "test")
	tokens := l.Tokenize()

	expected := []TokenType{
		TOKEN_IDENT, TOKEN_ASSIGN, TOKEN_INT, TOKEN_NEWLINE, TOKEN_EOF,
	}

	if len(tokens) != len(expected) {
		t.Fatalf("期望 %d 个 token，实际 %d", len(expected), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token[%d]: 期望 %v，实际 %v (%q)", i, expected[i], tok.Type, tok.Value)
		}
	}
}

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'world'`, "world"},
		{`"escape\n"`, "escape\n"},
		{`"tab\t"`, "tab\t"},
		{`"quote\""`, `quote"`},
	}

	for _, tt := range tests {
		l := New(tt.input, "test")
		tokens := l.Tokenize()
		if len(tokens) < 1 || tokens[0].Type != TOKEN_STRING {
			t.Errorf("输入 %q: 期望 STRING token", tt.input)
			continue
		}
		if tokens[0].Value != tt.expected {
			t.Errorf("输入 %q: 期望 %q，实际 %q", tt.input, tt.expected, tokens[0].Value)
		}
	}
}

func TestTripleQuoteStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"""hello"""`, "hello"},
		{`"""he"llo"""`, `he"llo`},
		// 三引号是原始字符串，\n 不转义
		{`"""line1\nline2"""`, `line1\nline2`},
	}

	for _, tt := range tests {
		l := New(tt.input, "test")
		tokens := l.Tokenize()
		found := false
		for _, tok := range tokens {
			if tok.Type == TOKEN_STRING {
				if tok.Value != tt.expected {
					t.Errorf("输入 %q: 期望 %q，实际 %q", tt.input, tt.expected, tok.Value)
				}
				if !tok.Raw {
					t.Errorf("输入 %q: 三引号字符串应标记为 Raw", tt.input)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("输入 %q: 未找到 STRING token", tt.input)
		}
	}
}

func TestShebang(t *testing.T) {
	input := "#!/usr/bin/env ops run\nprint('hello')"
	l := New(input, "test")
	tokens := l.Tokenize()

	// 第一行 shebang 应被跳过
	if tokens[0].Type != TOKEN_IDENT || tokens[0].Value != "print" {
		t.Errorf("shebang 未被正确跳过，第一个 token: %v %q", tokens[0].Type, tokens[0].Value)
	}
}

func TestIndentation(t *testing.T) {
	input := "if true\n    x = 1\ny = 2"
	l := New(input, "test")
	tokens := l.Tokenize()

	// 应该包含 INDENT 和 DEDENT
	hasIndent := false
	hasDedent := false
	for _, tok := range tokens {
		if tok.Type == TOKEN_INDENT {
			hasIndent = true
		}
		if tok.Type == TOKEN_DEDENT {
			hasDedent = true
		}
	}

	if !hasIndent {
		t.Error("缺少 INDENT token")
	}
	if !hasDedent {
		t.Error("缺少 DEDENT token")
	}
}

func TestOperators(t *testing.T) {
	input := "a == b != c <= d >= e && f || g"
	l := New(input, "test")
	tokens := l.Tokenize()

	expected := []TokenType{
		TOKEN_IDENT, TOKEN_EQ, TOKEN_IDENT, TOKEN_NE,
		TOKEN_IDENT, TOKEN_LE, TOKEN_IDENT, TOKEN_GE,
		TOKEN_IDENT, TOKEN_AND, TOKEN_IDENT, TOKEN_OR,
		TOKEN_IDENT, TOKEN_NEWLINE, TOKEN_EOF,
	}

	if len(tokens) != len(expected) {
		t.Fatalf("期望 %d 个 token，实际 %d", len(expected), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token[%d]: 期望 %v，实际 %v", i, expected[i], tok.Type)
		}
	}
}

func TestKeywords(t *testing.T) {
	input := "fn return if else for in while break continue"
	l := New(input, "test")
	tokens := l.Tokenize()

	expected := []TokenType{
		TOKEN_FN, TOKEN_RETURN, TOKEN_IF, TOKEN_ELSE,
		TOKEN_FOR, TOKEN_IN, TOKEN_WHILE, TOKEN_BREAK,
		TOKEN_CONTINUE, TOKEN_NEWLINE, TOKEN_EOF,
	}

	for i, tok := range tokens {
		if i >= len(expected) {
			break
		}
		if tok.Type != expected[i] {
			t.Errorf("token[%d]: 期望 %v，实际 %v (%q)", i, expected[i], tok.Type, tok.Value)
		}
	}
}
