// Package status keeps the latest rows shown in the status window.
package status

import (
	"sort"
	"sync"
)

// Store holds the latest rows and notifies a listener on change.
type Store struct {
	mu       sync.RWMutex
	rows     map[string]string
	onChange func()
}

// OnChange registers f, called after every Set (outside the lock). Only one listener exists.
func (s *Store) OnChange(f func()) {
	s.mu.Lock()
	s.onChange = f
	s.mu.Unlock()
}

// Set replaces the rows with a copy of rows.
func (s *Store) Set(rows map[string]string) {
	cp := make(map[string]string, len(rows))
	for k, v := range rows {
		cp[k] = v
	}
	s.mu.Lock()
	s.rows = cp
	f := s.onChange
	s.mu.Unlock()
	if f != nil {
		f()
	}
}

// Row is one table line.
type Row struct{ Key, Value string }

// Rows returns the rows sorted by key, the "Status" row first.
func (s *Store) Rows() []Row {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Row, 0, len(s.rows))
	for k, v := range s.rows {
		out = append(out, Row{k, v})
	}
	sort.Slice(out, func(i, j int) bool {
		if (out[i].Key == "Status") != (out[j].Key == "Status") {
			return out[i].Key == "Status"
		}
		return out[i].Key < out[j].Key
	})
	return out
}
