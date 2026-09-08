// Package state stores task lifecycle events and exposes current task state.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

type Status string

const (
	Planned     Status = "planned"
	Leased      Status = "leased"
	Started     Status = "started"
	Applied     Status = "applied"
	Verified    Status = "verified"
	Compensated Status = "compensated"
	Failed      Status = "failed"
)

type Event struct {
	ID      string    `json:"id"`
	TaskID  string    `json:"task_id"`
	Host    string    `json:"host"`
	Stage   string    `json:"stage"`
	Status  Status    `json:"status"`
	Message string    `json:"message,omitempty"`
	At      time.Time `json:"at"`
}

type Store struct {
	mu     sync.RWMutex
	events map[string]Event
}

// FileStore persists lifecycle events as one JSON document.
type FileStore struct {
	mu     sync.Mutex
	path   string
	events map[string]Event
}

func OpenFileStore(path string) (*FileStore, error) {
	if path == "" {
		return nil, fmt.Errorf("state store path is empty")
	}
	store := &FileStore{path: path, events: make(map[string]Event)}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read state store: %w", err)
	}
	if len(data) == 0 {
		return store, nil
	}
	if err := json.Unmarshal(data, &store.events); err != nil {
		return nil, fmt.Errorf("decode state store: %w", err)
	}
	if store.events == nil {
		store.events = make(map[string]Event)
	}
	return store, nil
}

func (s *FileStore) Append(event Event) error {
	if s == nil {
		return fmt.Errorf("file state store is nil")
	}
	if event.ID == "" || event.TaskID == "" || event.Host == "" || event.Stage == "" || event.Status == "" {
		return fmt.Errorf("state event requires id, task, host, stage, and status")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.events[event.ID]; ok {
		if event.At.IsZero() {
			event.At = existing.At
		}
		if !sameEvent(existing, event) {
			return fmt.Errorf("state event %q already exists with different content", event.ID)
		}
		return nil
	}
	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}
	s.events[event.ID] = event
	data, err := json.MarshalIndent(s.events, "", "  ")
	if err != nil {
		delete(s.events, event.ID)
		return fmt.Errorf("encode state store: %w", err)
	}
	temporary := s.path + ".tmp"
	if err := os.WriteFile(temporary, data, 0600); err != nil {
		delete(s.events, event.ID)
		return fmt.Errorf("write state store: %w", err)
	}
	if err := os.Rename(temporary, s.path); err != nil {
		delete(s.events, event.ID)
		return fmt.Errorf("commit state store: %w", err)
	}
	return nil
}

func (s *FileStore) Events(taskID, host string) []Event {
	if s == nil {
		return []Event{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Event, 0)
	for _, event := range s.events {
		if (taskID == "" || event.TaskID == taskID) && (host == "" || event.Host == host) {
			result = append(result, event)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

func NewStore() *Store { return &Store{events: make(map[string]Event)} }

func (s *Store) Append(event Event) error {
	if s == nil {
		return fmt.Errorf("state store is nil")
	}
	if event.ID == "" || event.TaskID == "" || event.Host == "" || event.Stage == "" || event.Status == "" {
		return fmt.Errorf("state event requires id, task, host, stage, and status")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.events[event.ID]; ok {
		if event.At.IsZero() {
			event.At = existing.At
		}
		if !sameEvent(existing, event) {
			return fmt.Errorf("state event %q already exists with different content", event.ID)
		}
		return nil
	}
	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}
	s.events[event.ID] = event
	return nil
}

func sameEvent(existing, incoming Event) bool {
	if incoming.At.IsZero() {
		incoming.At = existing.At
	}
	return existing == incoming
}

func (s *Store) Events(taskID, host string) []Event {
	if s == nil {
		return []Event{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Event, 0)
	for _, event := range s.events {
		if (taskID == "" || event.TaskID == taskID) && (host == "" || event.Host == host) {
			result = append(result, event)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}
