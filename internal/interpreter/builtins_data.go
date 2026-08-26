// interpreter/builtins_data.go: general-purpose data manipulation builtins
// (string, list, dict operations). These are language-level pure functions,
// NOT atomic ops: they live outside opsspec on purpose, so the runner
// instruction generator rejects them with its standard "unknown function"
// error (the linear runner VM has no expression evaluator). The interpreter
// and the AOT code generator implement identical semantics — sort() and
// keys() are deterministic in both so results never diverge.
package interpreter

import (
	"fmt"
	"sort"
	"strings"
)

// registerDataBuiltins installs the split/join/replace/upper/lower/trim/
// contains/index_of/sort/reverse/keys/values builtins.
func registerDataBuiltins(interp *Interpreter) {
	interp.builtins["split"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("split(s, sep) takes exactly 2 arguments, got %d", len(args))
		}
		s, ok1 := args[0].(string)
		sep, ok2 := args[1].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("split(s, sep) requires two strings, got %s and %s", typeName(args[0]), typeName(args[1]))
		}
		parts := strings.Split(s, sep)
		out := make([]interface{}, len(parts))
		for i, p := range parts {
			out[i] = p
		}
		return out, nil
	}

	interp.builtins["join"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("join(items, sep) takes exactly 2 arguments, got %d", len(args))
		}
		items, ok1 := args[0].([]interface{})
		sep, ok2 := args[1].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("join(items, sep) requires a list and a string, got %s and %s", typeName(args[0]), typeName(args[1]))
		}
		parts := make([]string, len(items))
		for i, item := range items {
			parts[i] = formatValue(item)
		}
		return strings.Join(parts, sep), nil
	}

	interp.builtins["replace"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("replace(s, old, new) takes exactly 3 arguments, got %d", len(args))
		}
		s, ok1 := args[0].(string)
		oldStr, ok2 := args[1].(string)
		newStr, ok3 := args[2].(string)
		if !ok1 || !ok2 || !ok3 {
			return nil, fmt.Errorf("replace(s, old, new) requires three strings, got %s, %s, %s", typeName(args[0]), typeName(args[1]), typeName(args[2]))
		}
		return strings.ReplaceAll(s, oldStr, newStr), nil
	}

	interp.builtins["upper"] = func(args ...interface{}) (interface{}, error) {
		s, err := oneString(args, "upper")
		if err != nil {
			return nil, err
		}
		return strings.ToUpper(s), nil
	}

	interp.builtins["lower"] = func(args ...interface{}) (interface{}, error) {
		s, err := oneString(args, "lower")
		if err != nil {
			return nil, err
		}
		return strings.ToLower(s), nil
	}

	interp.builtins["trim"] = func(args ...interface{}) (interface{}, error) {
		s, err := oneString(args, "trim")
		if err != nil {
			return nil, err
		}
		return strings.TrimSpace(s), nil
	}

	// contains works on all three container kinds: substring for strings,
	// element membership for lists (deep equality), key presence for dicts.
	interp.builtins["contains"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("contains(container, value) takes exactly 2 arguments, got %d", len(args))
		}
		switch c := args[0].(type) {
		case string:
			needle, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("contains(string, ...) requires a string needle, got %s", typeName(args[1]))
			}
			return strings.Contains(c, needle), nil
		case []interface{}:
			for _, item := range c {
				if isEqual(item, args[1]) {
					return true, nil
				}
			}
			return false, nil
		case map[string]interface{}:
			key, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("contains(dict, ...) requires a string key, got %s", typeName(args[1]))
			}
			_, exists := c[key]
			return exists, nil
		default:
			return nil, fmt.Errorf("contains() requires a string, list or dict, got %s", typeName(args[0]))
		}
	}

	// index_of returns the 0-based position of the first match, or -1.
	interp.builtins["index_of"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("index_of(container, value) takes exactly 2 arguments, got %d", len(args))
		}
		switch c := args[0].(type) {
		case string:
			needle, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("index_of(string, ...) requires a string needle, got %s", typeName(args[1]))
			}
			return int64(strings.Index(c, needle)), nil
		case []interface{}:
			for i, item := range c {
				if isEqual(item, args[1]) {
					return int64(i), nil
				}
			}
			return int64(-1), nil
		default:
			return nil, fmt.Errorf("index_of() requires a string or list, got %s", typeName(args[0]))
		}
	}

	// sort returns a NEW sorted list; the input is never mutated. All
	// numeric lists sort numerically, all-string lists lexicographically,
	// mixed types are an explicit error rather than a silent order.
	interp.builtins["sort"] = func(args ...interface{}) (interface{}, error) {
		items, err := oneList(args, "sort")
		if err != nil {
			return nil, err
		}
		out := make([]interface{}, len(items))
		copy(out, items)
		allNum, allStr := true, true
		for _, v := range out {
			switch v.(type) {
			case int64, float64:
				allStr = false
			case string:
				allNum = false
			default:
				allNum, allStr = false, false
			}
		}
		switch {
		case allNum:
			sort.SliceStable(out, func(i, j int) bool {
				return numOf(out[i]) < numOf(out[j])
			})
		case allStr:
			sort.SliceStable(out, func(i, j int) bool {
				return out[i].(string) < out[j].(string)
			})
		default:
			return nil, fmt.Errorf("sort() requires all numbers or all strings; mixed list cannot be ordered")
		}
		return out, nil
	}

	// reverse returns a NEW reversed list; the input is never mutated.
	interp.builtins["reverse"] = func(args ...interface{}) (interface{}, error) {
		items, err := oneList(args, "reverse")
		if err != nil {
			return nil, err
		}
		out := make([]interface{}, len(items))
		for i, v := range items {
			out[len(items)-1-i] = v
		}
		return out, nil
	}

	// keys returns dict keys sorted alphabetically. Go map iteration order
	// is random; sorting keeps interpreter and AOT output identical.
	interp.builtins["keys"] = func(args ...interface{}) (interface{}, error) {
		dict, err := oneDict(args, "keys")
		if err != nil {
			return nil, err
		}
		keysList := mapKeys(dict)
		sort.Strings(keysList)
		out := make([]interface{}, len(keysList))
		for i, k := range keysList {
			out[i] = k
		}
		return out, nil
	}

	// values returns dict values ordered by their sorted keys, matching
	// keys() exactly so zip-style pairing stays consistent.
	interp.builtins["values"] = func(args ...interface{}) (interface{}, error) {
		dict, err := oneDict(args, "values")
		if err != nil {
			return nil, err
		}
		keysList := mapKeys(dict)
		sort.Strings(keysList)
		out := make([]interface{}, len(keysList))
		for i, k := range keysList {
			out[i] = dict[k]
		}
		return out, nil
	}
}

// numOf extracts a float from an int64/float64 value. Only called after
// the allNum gate has verified every element is numeric.
func numOf(v interface{}) float64 {
	switch n := v.(type) {
	case int64:
		return float64(n)
	case float64:
		return n
	default:
		return 0
	}
}

// oneString enforces exactly-one-string-argument builtins.
func oneString(args []interface{}, name string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("%s() takes exactly 1 argument, got %d", name, len(args))
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("%s() requires a string, got %s", name, typeName(args[0]))
	}
	return s, nil
}

// oneList enforces exactly-one-list-argument builtins.
func oneList(args []interface{}, name string) ([]interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s() takes exactly 1 argument, got %d", name, len(args))
	}
	items, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("%s() requires a list, got %s", name, typeName(args[0]))
	}
	return items, nil
}

// oneDict enforces exactly-one-dict-argument builtins.
func oneDict(args []interface{}, name string) (map[string]interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s() takes exactly 1 argument, got %d", name, len(args))
	}
	dict, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s() requires a dict, got %s", name, typeName(args[0]))
	}
	return dict, nil
}
