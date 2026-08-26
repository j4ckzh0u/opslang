package docgen

import (
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/opsspec"
)

// TestGenerateCoversEveryOp pins the core promise of docs-as-code: the
// generated index mentions every operation in opsspec. If this fails,
// someone edited the spec without regenerating docs (or broke grouping).
func TestGenerateCoversEveryOp(t *testing.T) {
	doc, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !strings.Contains(doc, "请勿手改") {
		t.Error("generated doc must warn against hand edits")
	}
	missing := 0
	for _, name := range opsspec.Names(nil) {
		if !strings.Contains(doc, "`"+name+"`") {
			t.Errorf("generated index missing op: %s", name)
			missing++
			if missing >= 5 {
				t.Fatal("too many missing ops, stopping")
			}
		}
	}
}

// TestGenerateGroupsAndColumns checks structural invariants: package
// sections exist and every table row has four columns.
func TestGenerateStructure(t *testing.T) {
	doc, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	pkgs := map[string]bool{}
	for _, fn := range opsspec.Funcs {
		pkg := fn.Name[:strings.Index(fn.Name, ".")]
		if !pkgs[pkg] {
			pkgs[pkg] = true
			if !strings.Contains(doc, "## "+pkg+"\n") {
				t.Errorf("missing section for package %s", pkg)
			}
		}
	}
	for _, line := range strings.Split(doc, "\n") {
		if strings.HasPrefix(line, "| `") {
			cells := strings.Count(line, "|")
			if cells != 5 { // 4 columns => 5 pipes
				t.Errorf("row does not have 4 columns: %q", line)
				break
			}
		}
	}
}

// TestControllerOnlyScopeRendered verifies the scope column distinguishes
// controller-only operations.
func TestControllerOnlyScopeRendered(t *testing.T) {
	doc, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	hasControllerOnly := false
	for _, fn := range opsspec.Funcs {
		if fn.Avail == opsspec.ControllerOnly {
			hasControllerOnly = true
			break
		}
	}
	if hasControllerOnly && !strings.Contains(doc, "仅控制器") {
		t.Error("spec has controller-only ops but doc never says 仅控制器")
	}
}
