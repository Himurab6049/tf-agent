package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Record is one line in the JSONL session file.
type Record struct {
	Type      string          `json:"type"`
	Content   string          `json:"content,omitempty"`
	Model     string          `json:"model,omitempty"`
	Usage     *UsageRecord    `json:"usage,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// UsageRecord carries token counts.
type UsageRecord struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

// Store manages a single session JSONL file.
type Store struct {
	mu      sync.Mutex
	id      string
	path    string
	records []Record
}

// New creates or opens a session by ID. Pass "" to generate a new ID.
func New(sessionsDir, id string) (*Store, error) {
	if id == "" {
		id = uuid.New().String()
	}
	path := filepath.Join(sessionsDir, id+".jsonl")

	s := &Store{id: id, path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) ID() string { return s.id }

func (s *Store) load() error {
	f, err := os.Open(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var r Record
		if err := json.Unmarshal(scanner.Bytes(), &r); err == nil {
			s.records = append(s.records, r)
		}
	}
	return scanner.Err()
}

// Append adds a record and flushes it to disk immediately.
func (s *Store) Append(r Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	r.Timestamp = time.Now()
	s.records = append(s.records, r)
	return s.appendToFile(r)
}

func (s *Store) appendToFile(r Record) error {
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

// Records returns a copy of all session records.
func (s *Store) Records() []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Record, len(s.records))
	copy(out, s.records)
	return out
}

// Clear removes all records from memory (does not delete the file).
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = nil
}

// SessionInfo describes a stored session.
type SessionInfo struct {
	ID      string
	Title   string
	ModTime time.Time
}

// ListSessions returns all session files in dir as (id, title, modtime) sorted newest first.
func ListSessions(dir string) ([]SessionInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	var infos []SessionInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".jsonl" {
			continue
		}
		id := name[:len(name)-len(".jsonl")]

		fi, err := e.Info()
		if err != nil {
			continue
		}

		title := firstUserMessage(filepath.Join(dir, name))

		infos = append(infos, SessionInfo{
			ID:      id,
			Title:   title,
			ModTime: fi.ModTime(),
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].ModTime.After(infos[j].ModTime)
	})

	return infos, nil
}

// firstUserMessage reads the first "user" record from a JSONL file and returns
// its content truncated to 60 characters.
func firstUserMessage(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var r Record
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue
		}
		if r.Type == "user" && r.Content != "" {
			if utf8.RuneCountInString(r.Content) > 60 {
				runes := []rune(r.Content)
				return string(runes[:60])
			}
			return r.Content
		}
	}
	return ""
}

// ResumeSession loads an existing session by ID from dir.
func ResumeSession(dir string, id string) (*Store, error) {
	path := filepath.Join(dir, id+".jsonl")
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("resume session: session %q not found", id)
	}
	s := &Store{id: id, path: path}
	if err := s.load(); err != nil {
		return nil, fmt.Errorf("resume session: %w", err)
	}
	return s, nil
}
