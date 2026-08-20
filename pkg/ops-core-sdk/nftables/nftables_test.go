package nftables

import (
	"os/exec"
	"testing"
)

func hasNft() bool {
	_, err := exec.LookPath("nft")
	return err == nil
}

func TestAddTableEmpty(t *testing.T) {
	_, err := AddTable("", "")
	if err == nil {
		t.Fatal("expected error for empty family/name")
	}
	_, err = AddTable("ip", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	_, err = AddTable("", "test")
	if err == nil {
		t.Fatal("expected error for empty family")
	}
}

func TestDeleteTableEmpty(t *testing.T) {
	_, err := DeleteTable("", "")
	if err == nil {
		t.Fatal("expected error for empty family/name")
	}
}

func TestAddChainEmpty(t *testing.T) {
	_, err := AddChain("", "", "", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestDeleteChainEmpty(t *testing.T) {
	_, err := DeleteChain("", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestAddRuleEmpty(t *testing.T) {
	_, err := AddRule("", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestDeleteRuleEmpty(t *testing.T) {
	_, err := DeleteRule("", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestFlushChainEmpty(t *testing.T) {
	_, err := FlushChain("", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestFlushTableEmpty(t *testing.T) {
	_, err := FlushTable("", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestAddSetEmpty(t *testing.T) {
	_, err := AddSet("", "", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestDeleteSetEmpty(t *testing.T) {
	_, err := DeleteSet("", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestAddElementEmpty(t *testing.T) {
	_, err := AddElement("", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestDeleteElementEmpty(t *testing.T) {
	_, err := DeleteElement("", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestListTablesNotAvailable(t *testing.T) {
	if !hasNft() {
		_, err := ListTables()
		if err == nil {
			t.Fatal("expected error when nft not found")
		}
	}
}

func TestListRulesetNotAvailable(t *testing.T) {
	if !hasNft() {
		_, err := ListRuleset()
		if err == nil {
			t.Fatal("expected error when nft not found")
		}
	}
}

func TestFlushRulesetNotAvailable(t *testing.T) {
	if !hasNft() {
		_, err := FlushRuleset()
		if err == nil {
			t.Fatal("expected error when nft not found")
		}
	}
}

func TestExportNotAvailable(t *testing.T) {
	if !hasNft() {
		_, err := Export("json")
		if err == nil {
			t.Fatal("expected error when nft not found")
		}
	}
}
