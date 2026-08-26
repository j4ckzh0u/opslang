// Package docgen generates the operator-facing operation index directly
// from opsspec, the single source of truth. Hand-written docs rot; the
// generated index cannot: every build of `make docs` reflects exactly
// what the three engines accept.
package docgen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/j4ckzh0u/opslang/internal/opsspec"
)

// Generate renders the complete operation index as Markdown, grouped by
// top-level package in opsspec order.
func Generate() (string, error) {
	var b strings.Builder
	b.WriteString("# 原子操作总索引（自动生成）\n\n")
	b.WriteString("> **本文档由 `make docs` 从 internal/opsspec/spec.go 自动生成，请勿手改。**\n")
	b.WriteString("> 参数名即调用时的位置参数顺序；`可变` 表示该操作会改变系统状态，需要 admin 及以上权限。\n\n")

	groups := groupByPackage()
	pkgNames := make([]string, 0, len(groups))
	for name := range groups {
		pkgNames = append(pkgNames, name)
	}
	sort.Slice(pkgNames, func(i, j int) bool {
		return firstIndex(pkgNames[i]) < firstIndex(pkgNames[j])
	})

	b.WriteString(fmt.Sprintf("共 %d 个原子操作，%d 个包。\n\n## 目录\n\n",
		len(opsspec.Funcs), len(groups)))
	for _, pkg := range pkgNames {
		fmt.Fprintf(&b, "- [%s](#%s)（%d 个）\n", pkg, pkg, len(groups[pkg]))
	}

	for _, pkg := range pkgNames {
		fmt.Fprintf(&b, "\n## %s\n\n", pkg)
		b.WriteString("| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |\n")
		b.WriteString("|---|---|---|---|\n")
		for _, fn := range groups[pkg] {
			args := "-"
			if len(fn.Args) > 0 {
				args = "`" + strings.Join(fn.Args, "`, `") + "`"
			}
			mutating := ""
			if fn.Mutating {
				mutating = "✓"
			}
			scope := "全部引擎"
			if fn.Avail == opsspec.ControllerOnly {
				scope = "仅控制器"
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", fn.Name, args, mutating, scope)
		}
	}
	return b.String(), nil
}

// OpCount returns the total number of atomic operations in the spec.
func OpCount() int { return len(opsspec.Funcs) }

func groupByPackage() map[string][]opsspec.Func {
	groups := make(map[string][]opsspec.Func)
	for _, fn := range opsspec.Funcs {
		pkg := fn.Name
		if idx := strings.Index(fn.Name, "."); idx > 0 {
			pkg = fn.Name[:idx]
		}
		groups[pkg] = append(groups[pkg], fn)
	}
	return groups
}

// firstIndex returns the position of a package's first entry in the
// canonical Funcs table so generated sections keep spec order.
func firstIndex(pkg string) int {
	for i, fn := range opsspec.Funcs {
		if strings.HasPrefix(fn.Name, pkg+".") || fn.Name == pkg {
			return i
		}
	}
	return len(opsspec.Funcs)
}
