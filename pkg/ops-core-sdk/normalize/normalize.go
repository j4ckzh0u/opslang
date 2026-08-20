// Package normalize provides data normalization operations.
package normalize

import (
	"strings"
	"unicode"
)

// NormalizeResult is returned by all operations.
type NormalizeResult struct {
	Original string `json:"original"`
	Result   string `json:"result"`
	Error    string `json:"error,omitempty"`
}

// Lower converts to lowercase.
func Lower(value string) NormalizeResult {
	return NormalizeResult{Original: value, Result: strings.ToLower(value)}
}

// Upper converts to uppercase.
func Upper(value string) NormalizeResult {
	return NormalizeResult{Original: value, Result: strings.ToUpper(value)}
}

// Trim removes leading and trailing whitespace.
func Trim(value string) NormalizeResult {
	return NormalizeResult{Original: value, Result: strings.TrimSpace(value)}
}

// Slugify converts to a URL-friendly slug.
func Slugify(value string) NormalizeResult {
	var b strings.Builder
	prev := false
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prev = false
		} else if !prev && b.Len() > 0 {
			b.WriteRune('-')
			prev = true
		}
	}
	result := strings.TrimRight(b.String(), "-")
	return NormalizeResult{Original: value, Result: result}
}

// Title converts to title case.
func Title(value string) NormalizeResult {
	return NormalizeResult{Original: value, Result: strings.Title(value)}
}

// CamelCase converts to camelCase.
func CamelCase(value string) NormalizeResult {
	words := splitWords(value)
	if len(words) == 0 {
		return NormalizeResult{Original: value, Result: ""}
	}
	var b strings.Builder
	for i, w := range words {
		if i == 0 {
			b.WriteString(strings.ToLower(w))
		} else {
			if len(w) > 0 {
				b.WriteString(strings.ToUpper(w[:1]))
				b.WriteString(strings.ToLower(w[1:]))
			}
		}
	}
	return NormalizeResult{Original: value, Result: b.String()}
}

// SnakeCase converts to snake_case.
func SnakeCase(value string) NormalizeResult {
	words := splitWords(value)
	lower := make([]string, len(words))
	for i, w := range words {
		lower[i] = strings.ToLower(w)
	}
	return NormalizeResult{Original: value, Result: strings.Join(lower, "_")}
}

// KebabCase converts to kebab-case.
func KebabCase(value string) NormalizeResult {
	words := splitWords(value)
	lower := make([]string, len(words))
	for i, w := range words {
		lower[i] = strings.ToLower(w)
	}
	return NormalizeResult{Original: value, Result: strings.Join(lower, "-")}
}

func splitWords(s string) []string {
	var words []string
	var current strings.Builder
	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' || r == '.' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		} else if unicode.IsUpper(r) && current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
			current.WriteRune(r)
		} else {
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}
