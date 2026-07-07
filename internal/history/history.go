package history

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Entry records a single cover letter generation.
type Entry struct {
	ID         string    `json:"id"`
	ProfileID  string    `json:"profileId"`
	ProfileName string   `json:"profileName"`
	Body       string    `json:"body"`
	OutputPath string    `json:"outputPath"`
	Filename   string    `json:"filename"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Store is a disk-backed history store.
type Store struct {
	path string
}

func NewStore(dataDir string) *Store {
	return &Store{path: filepath.Join(dataDir, "history.json")}
}

// Load reads all history entries. Returns empty slice if file missing.
func (s *Store) Load() ([]Entry, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("read history: %w", err)
	}
	if len(b) == 0 {
		return []Entry{}, nil
	}
	var es []Entry
	if err := json.Unmarshal(b, &es); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}
	return es, nil
}

func (s *Store) save(es []Entry) error {
	b, err := json.MarshalIndent(es, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	if err := os.WriteFile(s.path, b, 0644); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}

// Add records a new generation and returns the entry.
func (s *Store) Add(e Entry) (Entry, error) {
	if e.ProfileID == "" {
		return Entry{}, fmt.Errorf("profileId is required")
	}
	es, err := s.Load()
	if err != nil {
		return Entry{}, err
	}
	id, err := newID()
	if err != nil {
		return Entry{}, err
	}
	e.ID = id
	e.CreatedAt = time.Now().UTC()
	es = append(es, e)
	if err := s.save(es); err != nil {
		return Entry{}, err
	}
	return e, nil
}

// List returns entries, optionally filtered by profileID, newest first.
func (s *Store) List(profileID string) ([]Entry, error) {
	es, err := s.Load()
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, e := range es {
		if profileID != "" && e.ProfileID != profileID {
			continue
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func newID() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
