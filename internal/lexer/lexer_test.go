package lexer

import (
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/lexer/token"
)

// helper to tokenize and assert success.
func mustTokenize(t *testing.T, source string) []token.Token {
	t.Helper()
	l := New(source, "")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize(%q) unexpected error: %v", source, err)
	}
	return tokens
}

// helper to assert a tokenize error.
func mustError(t *testing.T, source string) error {
	t.Helper()
	l := New(source, "")
	_, err := l.Tokenize()
	if err == nil {
		t.Fatalf("Tokenize(%q) expected error, got nil", source)
	}
	return err
}

func TestEmptyInput(t *testing.T) {
	tokens := mustTokenize(t, "")
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token (EOF), got %d", len(tokens))
	}
	if tokens[0].Type != token.EOF {
		t.Errorf("expected EOF, got %v", tokens[0].Type)
	}
}

func TestWhitespaceOnly(t *testing.T) {
	cases := []string{"", " ", "\t", "  \t  ", " \t \t "}
	for _, src := range cases {
		tokens := mustTokenize(t, src)
		if len(tokens) != 1 || tokens[0].Type != token.EOF {
			t.Errorf("input %q: expected only EOF, got %v", src, tokens)
		}
	}
}

func TestSingleNewline(t *testing.T) {
	tokens := mustTokenize(t, "\n")
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Type != token.NEWLINE {
		t.Errorf("expected NEWLINE, got %v", tokens[0].Type)
	}
	if tokens[1].Type != token.EOF {
		t.Errorf("expected EOF, got %v", tokens[1].Type)
	}
}

func TestAllOperatorsAndDelimiters(t *testing.T) {
	cases := []struct {
		input   string
		want    token.TokenType
		literal string
	}{
		{"+", token.PLUS, "+"},
		{"-", token.MINUS, "-"},
		{"*", token.STAR, "*"},
		{"%", token.PERCENT, "%"},
		{"==", token.EQ, "=="},
		{"!=", token.NEQ, "!="},
		{"<", token.LT, "<"},
		{">", token.GT, ">"},
		{"<=", token.LTE, "<="},
		{">=", token.GTE, ">="},
		{"=", token.ASSIGN, "="},
		{"&&", token.AND, "&&"},
		{"||", token.OR, "||"},
		{"!", token.NOT, "!"},
		{".", token.DOT, "."},
		{",", token.COMMA, ","},
		{":", token.COLON, ":"},
		{";", token.SEMICOLON, ";"},
		{"(", token.LPAREN, "("},
		{")", token.RPAREN, ")"},
		{"{", token.LBRACE, "{"},
		{"}", token.RBRACE, "}"},
		{"[", token.LBRACKET, "["},
		{"]", token.RBRACKET, "]"},
	}

	for _, tt := range cases {
		t.Run(tt.literal, func(t *testing.T) {
			tokens := mustTokenize(t, tt.input)
			if len(tokens) != 2 {
				t.Fatalf("expected 2 tokens (token + EOF), got %d", len(tokens))
			}
			if tokens[0].Type != tt.want {
				t.Errorf("type = %v, want %v", tokens[0].Type, tt.want)
			}
			if tokens[0].Literal != tt.literal {
				t.Errorf("literal = %q, want %q", tokens[0].Literal, tt.literal)
			}
		})
	}
}

func TestSlashAsOperator(t *testing.T) {
	// "/" alone (not followed by / or *) should be SLASH
	tokens := mustTokenize(t, "/ ")
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != token.SLASH {
		t.Errorf("type = %v, want SLASH", tokens[0].Type)
	}
	if tokens[0].Literal != "/" {
		t.Errorf("literal = %q, want %q", tokens[0].Literal, "/")
	}
}

func TestDoubleQuotedStrings(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"empty string", `""`, "", false},
		{"simple", `"hello"`, "hello", false},
		{"with spaces", `"hello world"`, "hello world", false},
		{"with digits", `"abc123"`, "abc123", false},
		{"newline escape", `"hello\nworld"`, "hello\nworld", false},
		{"tab escape", `"hello\tworld"`, "hello\tworld", false},
		{"carriage return escape", `"hello\rworld"`, "hello\rworld", false},
		{"backslash escape", `"hello\\world"`, "hello\\world", false},
		{"quote escape", `"hello\"world"`, "hello\"world", false},
		{"null escape", `"hello\0world"`, "hello\x00world", false},
		{"multiple escapes", `"a\nb\tc\\d\"e\0"`, "a\nb\tc\\d\"e\x00", false},
		{"unterminated", `"hello`, "", true},
		{"invalid escape", `"hello\x"`, "", true},
		{"escape at EOF", `"hello\`, "", true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mustError(t, tt.input)
				return
			}
			tokens := mustTokenize(t, tt.input)
			if tokens[0].Type != token.STRING {
				t.Errorf("type = %v, want STRING", tokens[0].Type)
			}
			if tokens[0].Literal != tt.want {
				t.Errorf("literal = %q, want %q", tokens[0].Literal, tt.want)
			}
		})
	}
}

func TestNewlineInString(t *testing.T) {
	err := mustError(t, "\"hello\nworld\"")
	lexErr, ok := err.(*LexError)
	if !ok {
		t.Fatalf("expected *LexError, got %T", err)
	}
	if !strings.Contains(lexErr.Msg, "newline in string") {
		t.Errorf("expected 'newline in string' error, got %q", lexErr.Msg)
	}
}

func TestBacktickStrings(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty backtick", "``", ""},
		{"simple", "`hello`", "hello"},
		{"with double quotes inside", "`he\"llo`", "he\"llo"},
		{"with backslash no processing", "`hello\\nworld`", "hello\\nworld"},
		{"with newline inside", "`hello\nworld`", "hello\nworld"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.input)
			if tokens[0].Type != token.STRING {
				t.Errorf("type = %v, want STRING", tokens[0].Type)
			}
			if tokens[0].Literal != tt.want {
				t.Errorf("literal = %q, want %q", tokens[0].Literal, tt.want)
			}
		})
	}
}

func TestUnterminatedBacktick(t *testing.T) {
	mustError(t, "`hello")
}

func TestSingleQuotedCharLiterals(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple char", "'a'", "a", false},
		{"digit char", "'5'", "5", false},
		{"newline escape", "'\\n'", "\n", false},
		{"tab escape", "'\\t'", "\t", false},
		{"backslash escape", "'\\\\'", "\\", false},
		{"single quote escape", "'\\''", "'", false},
		{"unknown escape", "'\\x'", "", true},
		{"unterminated", "'a", "", true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mustError(t, tt.input)
				return
			}
			tokens := mustTokenize(t, tt.input)
			if tokens[0].Type != token.STRING {
				t.Errorf("type = %v, want STRING", tokens[0].Type)
			}
			if tokens[0].Literal != tt.want {
				t.Errorf("literal = %q, want %q", tokens[0].Literal, tt.want)
			}
		})
	}
}

func TestNumberLiterals(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantTyp token.TokenType
		wantLit string
	}{
		{"zero", "0", token.INT, "0"},
		{"simple int", "123", token.INT, "123"},
		{"large int", "999999", token.INT, "999999"},
		{"simple float", "1.5", token.FLOAT, "1.5"},
		{"float with many decimals", "3.14159", token.FLOAT, "3.14159"},
		{"sci notation lowercase", "1e10", token.FLOAT, "1e10"},
		{"sci notation uppercase", "1E10", token.FLOAT, "1E10"},
		{"sci notation with dot", "1.5e10", token.FLOAT, "1.5e10"},
		{"sci notation negative exp", "1e-5", token.FLOAT, "1e-5"},
		{"sci notation positive exp", "1e+5", token.FLOAT, "1e+5"},
		{"sci notation with dot and neg exp", "3.14e-2", token.FLOAT, "3.14e-2"},
		{"leading zeros", "007", token.INT, "007"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.input)
			if tokens[0].Type != tt.wantTyp {
				t.Errorf("type = %v, want %v", tokens[0].Type, tt.wantTyp)
			}
			if tokens[0].Literal != tt.wantLit {
				t.Errorf("literal = %q, want %q", tokens[0].Literal, tt.wantLit)
			}
		})
	}
}

func TestNumberEdgeCases(t *testing.T) {
	// "1." should be INT "1" + DOT "." (no digit after dot)
	tokens := mustTokenize(t, "1.")
	if tokens[0].Type != token.INT || tokens[0].Literal != "1" {
		t.Errorf("first token: got %v %q, want INT \"1\"", tokens[0].Type, tokens[0].Literal)
	}
	if tokens[1].Type != token.DOT {
		t.Errorf("second token: got %v, want DOT", tokens[1].Type)
	}

	// "1.2.3" should be FLOAT "1.2" + DOT "." + INT "3"
	tokens = mustTokenize(t, "1.2.3")
	if tokens[0].Type != token.FLOAT || tokens[0].Literal != "1.2" {
		t.Errorf("first: got %v %q, want FLOAT \"1.2\"", tokens[0].Type, tokens[0].Literal)
	}
	if tokens[1].Type != token.DOT {
		t.Errorf("second: got %v, want DOT", tokens[1].Type)
	}
	if tokens[2].Type != token.INT || tokens[2].Literal != "3" {
		t.Errorf("third: got %v %q, want INT \"3\"", tokens[2].Type, tokens[2].Literal)
	}
}

func TestInvalidNumberExponent(t *testing.T) {
	cases := []string{"1e", "1e+", "1e-"}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			err := mustError(t, src)
			lexErr, ok := err.(*LexError)
			if !ok {
				t.Fatalf("expected *LexError, got %T", err)
			}
			if !strings.Contains(lexErr.Msg, "expected digit after exponent") {
				t.Errorf("expected exponent error, got %q", lexErr.Msg)
			}
		})
	}
}

func TestIdentifiersAndKeywords(t *testing.T) {
	cases := []struct {
		input   string
		wantTyp token.TokenType
		wantLit string
	}{
		{"foo", token.IDENT, "foo"},
		{"myVar", token.IDENT, "myVar"},
		{"x", token.IDENT, "x"},
		{"_", token.IDENT, "_"},
		{"_private", token.IDENT, "_private"},
		{"camelCase", token.IDENT, "camelCase"},
		{"PascalCase", token.IDENT, "PascalCase"},
		{"snake_case", token.IDENT, "snake_case"},
		{"with123", token.IDENT, "with123"},
		{"abc123def", token.IDENT, "abc123def"},
		// Keywords
		{"let", token.LET, "let"},
		{"fn", token.FN, "fn"},
		{"if", token.IF, "if"},
		{"else", token.ELSE, "else"},
		{"for", token.FOR, "for"},
		{"while", token.WHILE, "while"},
		{"return", token.RETURN, "return"},
		{"task", token.TASK, "task"},
		{"on", token.ON, "on"},
		{"import", token.IMPORT, "import"},
		{"report", token.REPORT, "report"},
		{"alert", token.ALERT, "alert"},
		// Boolean / nil literals (looked up as keywords)
		{"true", token.TRUE, "true"},
		{"false", token.FALSE, "false"},
		{"nil", token.NIL, "nil"},
	}

	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			tokens := mustTokenize(t, tt.input)
			if tokens[0].Type != tt.wantTyp {
				t.Errorf("type = %v, want %v", tokens[0].Type, tt.wantTyp)
			}
			if tokens[0].Literal != tt.wantLit {
				t.Errorf("literal = %q, want %q", tokens[0].Literal, tt.wantLit)
			}
		})
	}
}

func TestKeywordsAreCaseSensitive(t *testing.T) {
	// Uppercase versions of keywords should be identifiers, not keywords
	cases := []string{"LET", "IF", "FN", "Let", "If", "For", "Return"}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			tokens := mustTokenize(t, src)
			if tokens[0].Type != token.IDENT {
				t.Errorf("expected IDENT for %q, got %v", src, tokens[0].Type)
			}
		})
	}
}

func TestSingleLineComments(t *testing.T) {
	// Comment alone (followed by EOF)
	tokens := mustTokenize(t, "// this is a comment")
	if len(tokens) != 1 || tokens[0].Type != token.EOF {
		t.Errorf("comment-only: expected EOF, got %v", tokens)
	}

	// Comment followed by newline
	tokens = mustTokenize(t, "// comment\nlet x")
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.NEWLINE, "\n"},
		{token.LET, "let"},
		{token.IDENT, "x"},
		{token.EOF, ""},
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i, exp := range expected {
		if tokens[i].Type != exp.typ {
			t.Errorf("token[%d] type = %v, want %v", i, tokens[i].Type, exp.typ)
		}
		if tokens[i].Literal != exp.lit {
			t.Errorf("token[%d] literal = %q, want %q", i, tokens[i].Literal, exp.lit)
		}
	}

	// Code then comment
	tokens = mustTokenize(t, "let x // comment")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens (let, x, EOF), got %d", len(tokens))
	}
	if tokens[0].Type != token.LET || tokens[1].Type != token.IDENT {
		t.Errorf("expected LET IDENT, got %v %v", tokens[0].Type, tokens[1].Type)
	}
}

func TestBlockComments(t *testing.T) {
	// Simple block comment
	tokens := mustTokenize(t, "/* comment */")
	if len(tokens) != 1 || tokens[0].Type != token.EOF {
		t.Errorf("block comment alone: expected EOF, got %v", tokens)
	}

	// Block comment with content inside
	tokens = mustTokenize(t, "/* multi\nline\ncomment */")
	if len(tokens) != 1 || tokens[0].Type != token.EOF {
		t.Errorf("multi-line block comment: expected EOF, got %v", tokens)
	}

	// Code around block comment
	tokens = mustTokenize(t, "let /* skip */ x")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Type != token.LET {
		t.Errorf("expected LET, got %v", tokens[0].Type)
	}
	if tokens[1].Type != token.IDENT || tokens[1].Literal != "x" {
		t.Errorf("expected IDENT 'x', got %v %q", tokens[1].Type, tokens[1].Literal)
	}

	// Nested-looking block comment (/* a /* b */ c */)
	// The first */ closes the comment, so "c */" would be parsed as code.
	// After "/* a /* b */" the remaining is " c */", which has "c" (IDENT) then error on "*".
	// Actually, let's not test nested — it's not supported. Just test that the first */ closes.
	tokens = mustTokenize(t, "/* a /* b */ let x")
	expectedTypes := []token.TokenType{token.LET, token.IDENT, token.EOF}
	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expectedTypes), len(tokens), tokens)
	}
	for i, exp := range expectedTypes {
		if tokens[i].Type != exp {
			t.Errorf("token[%d] type = %v, want %v", i, tokens[i].Type, exp)
		}
	}
}

func TestUnterminatedBlockComment(t *testing.T) {
	cases := []string{"/* unterminated", "/* still going", "/*"}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			err := mustError(t, src)
			lexErr, ok := err.(*LexError)
			if !ok {
				t.Fatalf("expected *LexError, got %T", err)
			}
			if !strings.Contains(lexErr.Msg, "unterminated block comment") {
				t.Errorf("expected 'unterminated block comment' error, got %q", lexErr.Msg)
			}
		})
	}
}

func TestConsecutiveComments(t *testing.T) {
	tokens := mustTokenize(t, "// first\n// second\nlet x")
	expected := []token.TokenType{token.NEWLINE, token.NEWLINE, token.LET, token.IDENT, token.EOF}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("token[%d] type = %v, want %v", i, tokens[i].Type, exp)
		}
	}
}

func TestErrorCases(t *testing.T) {
	t.Run("unexpected backslash", func(t *testing.T) {
		err := mustError(t, "\\")
		lexErr, ok := err.(*LexError)
		if !ok {
			t.Fatalf("expected *LexError, got %T", err)
		}
		if !strings.Contains(lexErr.Msg, "unexpected character") {
			t.Errorf("expected 'unexpected character' error, got %q", lexErr.Msg)
		}
	})

	t.Run("lone ampersand", func(t *testing.T) {
		err := mustError(t, "&")
		lexErr, ok := err.(*LexError)
		if !ok {
			t.Fatalf("expected *LexError, got %T", err)
		}
		if !strings.Contains(lexErr.Msg, "did you mean '&&'") {
			t.Errorf("expected 'did you mean &&' error, got %q", lexErr.Msg)
		}
	})

	t.Run("lone pipe", func(t *testing.T) {
		err := mustError(t, "|")
		lexErr, ok := err.(*LexError)
		if !ok {
			t.Fatalf("expected *LexError, got %T", err)
		}
		if !strings.Contains(lexErr.Msg, "did you mean '||'") {
			t.Errorf("expected 'did you mean ||' error, got %q", lexErr.Msg)
		}
	})

	t.Run("at sign", func(t *testing.T) {
		err := mustError(t, "@")
		lexErr, ok := err.(*LexError)
		if !ok {
			t.Fatalf("expected *LexError, got %T", err)
		}
		if !strings.Contains(lexErr.Msg, "unexpected character") {
			t.Errorf("expected 'unexpected character' error, got %q", lexErr.Msg)
		}
	})

	t.Run("hash sign", func(t *testing.T) {
		err := mustError(t, "#")
		lexErr, ok := err.(*LexError)
		if !ok {
			t.Fatalf("expected *LexError, got %T", err)
		}
		if !strings.Contains(lexErr.Msg, "unexpected character") {
			t.Errorf("expected 'unexpected character' error, got %q", lexErr.Msg)
		}
	})
}

func TestPositionTracking(t *testing.T) {
	// Line 1: let x = 5
	// Line 2: let y = 10
	source := "let x = 5\nlet y = 10"
	tokens := mustTokenize(t, source)

	// Expected positions (line, col):
	// let  -> (1,1)
	// x    -> (1,5)
	// =    -> (1,7)
	// 5    -> (1,9)
	// \n   -> (1,10)
	// let  -> (2,1)
	// y    -> (2,5)
	// =    -> (2,7)
	// 10   -> (2,9)
	// EOF  -> (2,11)
	expected := []struct {
		lit  string
		line int
		col  int
	}{
		{"let", 1, 1},
		{"x", 1, 5},
		{"=", 1, 7},
		{"5", 1, 9},
		{"\n", 1, 10},
		{"let", 2, 1},
		{"y", 2, 5},
		{"=", 2, 7},
		{"10", 2, 9},
		{"", 2, 11}, // EOF
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, exp := range expected {
		if tokens[i].Pos.Line != exp.line || tokens[i].Pos.Column != exp.col {
			t.Errorf("token[%d] %q: pos = (%d:%d), want (%d:%d)",
				i, exp.lit, tokens[i].Pos.Line, tokens[i].Pos.Column, exp.line, exp.col)
		}
	}
}

func TestPositionTrackingWithFilename(t *testing.T) {
	l := New("let x", "test.ops")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens[0].Pos.File != "test.ops" {
		t.Errorf("expected filename 'test.ops', got %q", tokens[0].Pos.File)
	}
}

func TestPositionAfterBlockComment(t *testing.T) {
	// "/* abc */let x" -- block comment doesn't emit tokens, let starts after
	source := "/* abc */let x"
	tokens := mustTokenize(t, source)
	// "/* abc */" is 9 characters, so "let" starts at col 10
	if tokens[0].Pos.Line != 1 || tokens[0].Pos.Column != 10 {
		t.Errorf("let after block comment: pos = (%d:%d), want (1:10)",
			tokens[0].Pos.Line, tokens[0].Pos.Column)
	}
}

func TestUTF8Identifiers(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantTyp token.TokenType
		wantLit string
	}{
		{"chinese", "变量", token.IDENT, "变量"},
		{"japanese", "名前", token.IDENT, "名前"},
		{"korean", "변수", token.IDENT, "변수"},
		{"greek alpha", "αβγ", token.IDENT, "αβγ"},
		{"german umlaut", "größe", token.IDENT, "größe"},
		{"underscore unicode", "_变量", token.IDENT, "_变量"},
		{"unicode with digits", "变量123", token.IDENT, "变量123"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.input)
			if tokens[0].Type != tt.wantTyp {
				t.Errorf("type = %v, want %v", tokens[0].Type, tt.wantTyp)
			}
			if tokens[0].Literal != tt.wantLit {
				t.Errorf("literal = %q, want %q", tokens[0].Literal, tt.wantLit)
			}
		})
	}
}

func TestUTF8IdentifiersInExpression(t *testing.T) {
	// "let 变量 = 5"
	source := "let 变量 = 5"
	tokens := mustTokenize(t, source)

	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.LET, "let"},
		{token.IDENT, "变量"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.EOF, ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.typ {
			t.Errorf("token[%d] type = %v, want %v (literal %q)", i, tokens[i].Type, exp.typ, tokens[i].Literal)
		}
		if tokens[i].Literal != exp.lit {
			t.Errorf("token[%d] literal = %q, want %q", i, tokens[i].Literal, exp.lit)
		}
	}
}

func TestUTF8IdentifierPosition(t *testing.T) {
	// Chinese chars are multi-byte but each counts as 1 column
	// "变量 x" -- '变' at col 1, '量' at col 2, ' ' at col 3, 'x' at col 4
	source := "变量 x"
	tokens := mustTokenize(t, source)

	if tokens[0].Pos.Line != 1 || tokens[0].Pos.Column != 1 {
		t.Errorf("变量: pos = (%d:%d), want (1:1)", tokens[0].Pos.Line, tokens[0].Pos.Column)
	}
	if tokens[1].Pos.Line != 1 || tokens[1].Pos.Column != 4 {
		t.Errorf("x: pos = (%d:%d), want (1:4)", tokens[1].Pos.Line, tokens[1].Pos.Column)
	}
}

func TestComplexExpression(t *testing.T) {
	// "let result = (1 + 2) * 3.14"
	source := "let result = (1 + 2) * 3.14"
	tokens := mustTokenize(t, source)

	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.LPAREN, "("},
		{token.INT, "1"},
		{token.PLUS, "+"},
		{token.INT, "2"},
		{token.RPAREN, ")"},
		{token.STAR, "*"},
		{token.FLOAT, "3.14"},
		{token.EOF, ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.typ {
			t.Errorf("token[%d] type = %v, want %v", i, tokens[i].Type, exp.typ)
		}
		if tokens[i].Literal != exp.lit {
			t.Errorf("token[%d] literal = %q, want %q", i, tokens[i].Literal, exp.lit)
		}
	}
}

func TestFunctionDefinition(t *testing.T) {
	source := "fn add(a, b) { return a + b }"
	tokens := mustTokenize(t, source)

	expectedTypes := []token.TokenType{
		token.FN, token.IDENT, token.LPAREN, token.IDENT, token.COMMA,
		token.IDENT, token.RPAREN, token.LBRACE, token.RETURN,
		token.IDENT, token.PLUS, token.IDENT, token.RBRACE, token.EOF,
	}

	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d", len(expectedTypes), len(tokens))
	}

	for i, exp := range expectedTypes {
		if tokens[i].Type != exp {
			t.Errorf("token[%d] type = %v, want %v (literal %q)", i, tokens[i].Type, exp, tokens[i].Literal)
		}
	}
}

func TestTaskDeclaration(t *testing.T) {
	source := `task "deploy" on hosts {
    report { status: "ok" }
}`
	tokens := mustTokenize(t, source)

	// Verify first few tokens
	if tokens[0].Type != token.TASK || tokens[0].Literal != "task" {
		t.Errorf("first token: got %v %q, want TASK", tokens[0].Type, tokens[0].Literal)
	}
	if tokens[1].Type != token.STRING || tokens[1].Literal != "deploy" {
		t.Errorf("second token: got %v %q, want STRING \"deploy\"", tokens[1].Type, tokens[1].Literal)
	}
	if tokens[2].Type != token.ON || tokens[2].Literal != "on" {
		t.Errorf("third token: got %v %q, want ON", tokens[2].Type, tokens[2].Literal)
	}
	if tokens[3].Type != token.IDENT || tokens[3].Literal != "hosts" {
		t.Errorf("fourth token: got %v %q, want IDENT \"hosts\"", tokens[3].Type, tokens[3].Literal)
	}
}

func TestLexErrorMessageFormat(t *testing.T) {
	l := New("let x = @", "test.ops")
	_, err := l.Tokenize()
	if err == nil {
		t.Fatal("expected error")
	}

	lexErr, ok := err.(*LexError)
	if !ok {
		t.Fatalf("expected *LexError, got %T", err)
	}

	// Error message should include file:line:col
	msg := lexErr.Error()
	if !strings.Contains(msg, "test.ops") {
		t.Errorf("error message should contain filename: %q", msg)
	}
	if !strings.Contains(msg, "unexpected character") {
		t.Errorf("error message should describe the error: %q", msg)
	}
}

func TestMixedContent(t *testing.T) {
	// A realistic script snippet
	source := `// Get system info
let cpu = sys.cpu.usage()
if cpu.percent > 90 {
    alert("CPU high: " + str(cpu.percent))
}`
	tokens, err := New(source, "").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it doesn't error and produces sensible tokens
	if len(tokens) < 10 {
		t.Errorf("expected at least 10 tokens, got %d", len(tokens))
	}

	// Last token should be EOF
	if tokens[len(tokens)-1].Type != token.EOF {
		t.Errorf("last token should be EOF, got %v", tokens[len(tokens)-1].Type)
	}

	// Find the "if" keyword
	foundIf := false
	for _, tok := range tokens {
		if tok.Type == token.IF {
			foundIf = true
			break
		}
	}
	if !foundIf {
		t.Error("expected to find IF keyword in tokens")
	}
}

func TestIdentAfterNumber(t *testing.T) {
	// "123abc" should be INT "123" + IDENT "abc"
	tokens := mustTokenize(t, "123abc")
	if tokens[0].Type != token.INT || tokens[0].Literal != "123" {
		t.Errorf("first: got %v %q, want INT \"123\"", tokens[0].Type, tokens[0].Literal)
	}
	if tokens[1].Type != token.IDENT || tokens[1].Literal != "abc" {
		t.Errorf("second: got %v %q, want IDENT \"abc\"", tokens[1].Type, tokens[1].Literal)
	}
}

func TestNumberFollowedByDot(t *testing.T) {
	// "42.method" should be INT "42" + DOT "." + IDENT "method"
	tokens := mustTokenize(t, "42.method")
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.INT, "42"},
		{token.DOT, "."},
		{token.IDENT, "method"},
		{token.EOF, ""},
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, exp := range expected {
		if tokens[i].Type != exp.typ {
			t.Errorf("token[%d] type = %v, want %v", i, tokens[i].Type, exp.typ)
		}
		if tokens[i].Literal != exp.lit {
			t.Errorf("token[%d] literal = %q, want %q", i, tokens[i].Literal, exp.lit)
		}
	}
}

func TestMultipleStatements(t *testing.T) {
	source := "let a = 1\nlet b = 2\nlet c = a + b"
	tokens, err := New(source, "").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count the LET tokens
	letCount := 0
	for _, tok := range tokens {
		if tok.Type == token.LET {
			letCount++
		}
	}
	if letCount != 3 {
		t.Errorf("expected 3 LET tokens, got %d", letCount)
	}
}

func TestEOFPosition(t *testing.T) {
	tokens := mustTokenize(t, "abc")
	eof := tokens[len(tokens)-1]
	if eof.Type != token.EOF {
		t.Fatalf("last token should be EOF, got %v", eof.Type)
	}
	// After "abc" (3 chars), EOF should be at col 4
	if eof.Pos.Column != 4 {
		t.Errorf("EOF col = %d, want 4", eof.Pos.Column)
	}
	if eof.Pos.Line != 1 {
		t.Errorf("EOF line = %d, want 1", eof.Pos.Line)
	}
}
