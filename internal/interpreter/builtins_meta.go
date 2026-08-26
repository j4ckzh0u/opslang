// interpreter/builtins_meta.go: introspection builtins over opsspec.
// doc(op) returns one operation's signature; ops(prefix) lists known
// operations. Like the data builtins these are controller-side language
// functions, NOT atomic ops: they live outside opsspec on purpose so
// runner-mode instruction generation rejects them (a remote linear VM has
// no registry to introspect), and the AOT compiler refuses them for the
// same reason — a compiled binary carries no spec table.
package interpreter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/j4ckzh0u/opslang/internal/opsspec"
)

// registerMetaBuiltins installs the doc/ops introspection builtins.
func registerMetaBuiltins(interp *Interpreter) {
	// doc("apt.install") -> {name, args, mutating, scope}
	interp.builtins["doc"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("doc(op) takes exactly 1 argument (operation name), got %d", len(args))
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("doc(op) requires a string operation name, got %s", typeName(args[0]))
		}
		// Historical aliases resolve to their canonical name so docs stay
		// consistent with what generators emit.
		if mapped, isAlias := opsspec.Aliases[name]; isAlias {
			name = mapped
		}
		fn, known := opsspec.Lookup(name)
		if !known {
			return nil, fmt.Errorf("doc(): unknown operation %q - use ops() to list all operations", name)
		}
		scope := "all engines"
		if fn.Avail == opsspec.ControllerOnly {
			scope = "controller only"
		}
		argList := make([]interface{}, 0, len(fn.Args))
		for _, a := range fn.Args {
			argList = append(argList, a)
		}
		out := map[string]interface{}{
			"name":     fn.Name,
			"args":     argList,
			"mutating": fn.Mutating,
			"scope":    scope,
		}
		return out, nil
	}

	// ops() / ops("sys.") -> sorted operation names, optionally by prefix
	interp.builtins["ops"] = func(args ...interface{}) (interface{}, error) {
		if len(args) > 1 {
			return nil, fmt.Errorf("ops(prefix) takes at most 1 argument, got %d", len(args))
		}
		prefix := ""
		if len(args) == 1 {
			p, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("ops(prefix) requires a string prefix, got %s", typeName(args[0]))
			}
			prefix = p
		}
		names := opsspec.Names(nil)
		sort.Strings(names)
		out := make([]interface{}, 0, len(names))
		for _, n := range names {
			if strings.HasPrefix(n, prefix) {
				out = append(out, n)
			}
		}
		return out, nil
	}
}
