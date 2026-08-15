package token

import "testing"

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		tt       TokenType
		expected string
	}{
		{ILLEGAL, "ILLEGAL"},
		{EOF, "EOF"},
		{NEWLINE, "NEWLINE"},
		{IDENT, "IDENT"},
		{INT, "INT"},
		{FLOAT, "FLOAT"},
		{STRING, "STRING"},
		{TRUE, "TRUE"},
		{FALSE, "FALSE"},
		{NIL, "NIL"},
		{LET, "LET"},
		{FN, "FN"},
		{IF, "IF"},
		{ELSE, "ELSE"},
		{FOR, "FOR"},
		{WHILE, "WHILE"},
		{RETURN, "RETURN"},
		{TASK, "TASK"},
		{ON, "ON"},
		{IMPORT, "IMPORT"},
		{REPORT, "REPORT"},
		{ALERT, "ALERT"},
		{PLUS, "PLUS"},
		{MINUS, "MINUS"},
		{STAR, "STAR"},
		{SLASH, "SLASH"},
		{ASSIGN, "ASSIGN"},
		{EQ, "EQ"},
		{LPAREN, "LPAREN"},
		{RPAREN, "RPAREN"},
		{LBRACE, "LBRACE"},
		{RBRACE, "RBRACE"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.tt.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestTokenTypeStringUnknown(t *testing.T) {
	tt := TokenType(9999)
	got := tt.String()
	if got != "TokenType(9999)" {
		t.Errorf("expected 'TokenType(9999)', got %q", got)
	}
}

func TestPositionString(t *testing.T) {
	tests := []struct {
		pos      Position
		expected string
	}{
		{Position{Line: 1, Column: 5, File: "test.ops"}, "test.ops:1:5"},
		{Position{Line: 10, Column: 20, File: ""}, "10:20"},
		{Position{Line: 1, Column: 1, File: "main.ops"}, "main.ops:1:1"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.pos.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestTokenString(t *testing.T) {
	tok := Token{
		Type:    INT,
		Literal: "42",
		Pos:     Position{Line: 1, Column: 5, File: "test.ops"},
	}
	got := tok.String()
	expected := `Token{INT, "42", test.ops:1:5}`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestLookupKeyword(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"let", LET},
		{"fn", FN},
		{"if", IF},
		{"else", ELSE},
		{"for", FOR},
		{"while", WHILE},
		{"return", RETURN},
		{"task", TASK},
		{"on", ON},
		{"import", IMPORT},
		{"true", TRUE},
		{"false", FALSE},
		{"nil", NIL},
		{"report", REPORT},
		{"alert", ALERT},
		{"foo", IDENT},
		{"x", IDENT},
		{"unknown", IDENT},
		{"", IDENT},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := LookupKeyword(tt.input)
			if got != tt.expected {
				t.Errorf("LookupKeyword(%q) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsKeyword(t *testing.T) {
	keywords := []string{"let", "fn", "if", "else", "for", "while", "return",
		"task", "on", "import", "true", "false", "nil", "report", "alert"}
	for _, kw := range keywords {
		t.Run(kw, func(t *testing.T) {
			if !IsKeyword(kw) {
				t.Errorf("IsKeyword(%q) = false, want true", kw)
			}
		})
	}

	nonKeywords := []string{"foo", "x", "print", "unknown", ""}
	for _, nk := range nonKeywords {
		t.Run("non_"+nk, func(t *testing.T) {
			if IsKeyword(nk) {
				t.Errorf("IsKeyword(%q) = true, want false", nk)
			}
		})
	}
}
