package zfs

import (
	"encoding/json"
	"testing"
)

func TestCreate_EmptyName(t *testing.T) {
	_, err := Create("", nil)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDestroy_EmptyName(t *testing.T) {
	_, err := Destroy("", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSet_EmptyArgs(t *testing.T) {
	_, err := Set("", "prop", "val")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	_, err = Set("ds", "", "val")
	if err == nil {
		t.Fatal("expected error for empty property")
	}
}

func TestGet_EmptyName(t *testing.T) {
	_, err := Get("", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestExists_EmptyName(t *testing.T) {
	_, err := Exists("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSnapshot_EmptyArgs(t *testing.T) {
	_, err := Snapshot("", "snap")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	_, err = Snapshot("ds", "")
	if err == nil {
		t.Fatal("expected error for empty snapshot name")
	}
}

func TestDestroySnapshot_EmptyArgs(t *testing.T) {
	_, err := DestroySnapshot("", "snap")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	_, err = DestroySnapshot("ds", "")
	if err == nil {
		t.Fatal("expected error for empty snapshot name")
	}
}

func TestActionResultJSON(t *testing.T) {
	r := ActionResult{Name: "tank/data", Changed: true, Success: true}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestDatasetInfoJSON(t *testing.T) {
	ds := DatasetInfo{
		Name:       "tank/data",
		Type:       "filesystem",
		Used:       "1G",
		Avail:      "10G",
		Refer:      "1G",
		Mountpoint: "/tank/data",
	}
	b, err := json.Marshal(ds)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestPoolInfoJSON(t *testing.T) {
	pool := PoolInfo{
		Name:   "tank",
		Size:   "100G",
		Used:   "50G",
		Avail:  "50G",
		Health: "ONLINE",
	}
	b, err := json.Marshal(pool)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestListResultJSON(t *testing.T) {
	result := ListResult{
		Datasets: []DatasetInfo{
			{Name: "tank/data", Type: "filesystem"},
		},
	}
	b, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestPoolListResultJSON(t *testing.T) {
	result := PoolListResult{
		Pools: []PoolInfo{
			{Name: "tank", Health: "ONLINE"},
		},
	}
	b, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}
