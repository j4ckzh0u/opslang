package state

import (
	"path/filepath"
	"testing"
)

func TestStoreAppendIdempotent(t *testing.T) {
	store := NewStore()
	event := Event{ID: "e1", TaskID: "t1", Host: "h1", Stage: "apply", Status: Applied}
	if err := store.Append(event); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(event); err != nil {
		t.Fatal(err)
	}
	if got := store.Events("t1", "h1"); len(got) != 1 {
		t.Fatalf("events = %d, want 1", len(got))
	}
	if err := store.Append(Event{ID: "e1", TaskID: "t1", Host: "h1", Stage: "verify", Status: Verified}); err == nil {
		t.Fatal("expected conflicting event error")
	}
}

func TestStoreRejectsEmptyEvent(t *testing.T) {
	if err := NewStore().Append(Event{}); err == nil {
		t.Fatal("expected validation error")
	}
	if got := (*Store)(nil).Events("", ""); got == nil {
		t.Fatal("nil store must return an empty slice")
	}
}

func TestStoreEventsAreStable(t *testing.T) {
	store := NewStore()
	for _, event := range []Event{
		{ID: "b", TaskID: "t", Host: "h", Stage: "apply", Status: Applied},
		{ID: "a", TaskID: "t", Host: "h", Stage: "plan", Status: Planned},
	} {
		event.At = event.At.UTC()
		if err := store.Append(event); err != nil {
			t.Fatal(err)
		}
	}
	events := store.Events("t", "h")
	if len(events) != 2 || events[0].ID != "a" || events[1].ID != "b" {
		t.Fatalf("events = %#v, want sorted IDs", events)
	}
}

func TestFileStorePersistsAndIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	store, err := OpenFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	event := Event{ID: "e1", TaskID: "t1", Host: "h1", Stage: "apply", Status: Applied}
	if err := store.Append(event); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(event); err != nil {
		t.Fatal(err)
	}
	reloaded, err := OpenFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.Events("t1", "h1"); len(got) != 1 || got[0].ID != "e1" {
		t.Fatalf("reloaded events = %#v", got)
	}
}
