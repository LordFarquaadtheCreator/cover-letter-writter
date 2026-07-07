package profile

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

// Profile holds a user's contact info used to generate cover letters.
type Profile struct {
	ID        string    `json:"id"`
	Label     string    `json:"label,omitempty"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	OutputDir string    `json:"outputDir,omitempty"`
	Filename  string    `json:"filename,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Store is a disk-backed profile store. The file is loaded on every call so
// edits made outside the process are visible — mirrors the model_card.yaml
// pattern in create-image.
type Store struct {
	path string
}

func NewStore(dataDir string) *Store {
	return &Store{path: filepath.Join(dataDir, "profiles.json")}
}

// Load reads all profiles from disk. Returns empty slice if file missing.
func (s *Store) Load() ([]Profile, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Profile{}, nil
		}
		return nil, fmt.Errorf("read profiles: %w", err)
	}
	if len(b) == 0 {
		return []Profile{}, nil
	}
	var ps []Profile
	if err := json.Unmarshal(b, &ps); err != nil {
		return nil, fmt.Errorf("parse profiles: %w", err)
	}
	return ps, nil
}

func (s *Store) save(ps []Profile) error {
	b, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal profiles: %w", err)
	}
	if err := os.WriteFile(s.path, b, 0644); err != nil {
		return fmt.Errorf("write profiles: %w", err)
	}
	return nil
}

// Create adds a new profile and returns it.
func (s *Store) Create(p Profile) (Profile, error) {
	if p.Name == "" {
		return Profile{}, fmt.Errorf("name is required")
	}
	if p.Email == "" {
		return Profile{}, fmt.Errorf("email is required")
	}
	ps, err := s.Load()
	if err != nil {
		return Profile{}, err
	}
	id, err := newID()
	if err != nil {
		return Profile{}, err
	}
	now := time.Now().UTC()
	p.ID = id
	p.CreatedAt = now
	p.UpdatedAt = now
	ps = append(ps, p)
	if err := s.save(ps); err != nil {
		return Profile{}, err
	}
	return p, nil
}

// Get returns a single profile by ID.
func (s *Store) Get(id string) (Profile, error) {
	ps, err := s.Load()
	if err != nil {
		return Profile{}, err
	}
	for _, p := range ps {
		if p.ID == id {
			return p, nil
		}
	}
	return Profile{}, fmt.Errorf("profile %q not found", id)
}

// List returns all profiles sorted by CreatedAt ascending.
func (s *Store) List() ([]Profile, error) {
	ps, err := s.Load()
	if err != nil {
		return nil, err
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].CreatedAt.Before(ps[j].CreatedAt) })
	return ps, nil
}

// Update applies non-zero fields of patch to the profile with id.
func (s *Store) Update(id string, patch Profile) (Profile, error) {
	ps, err := s.Load()
	if err != nil {
		return Profile{}, err
	}
	for i, p := range ps {
		if p.ID != id {
			continue
		}
		if patch.Label != "" {
			p.Label = patch.Label
		}
		if patch.Name != "" {
			p.Name = patch.Name
		}
		if patch.Address != "" {
			p.Address = patch.Address
		}
		if patch.Email != "" {
			p.Email = patch.Email
		}
		if patch.Phone != "" {
			p.Phone = patch.Phone
		}
		if patch.OutputDir != "" {
			p.OutputDir = patch.OutputDir
		}
		if patch.Filename != "" {
			p.Filename = patch.Filename
		}
		p.UpdatedAt = time.Now().UTC()
		ps[i] = p
		if err := s.save(ps); err != nil {
			return Profile{}, err
		}
		return p, nil
	}
	return Profile{}, fmt.Errorf("profile %q not found", id)
}

// Delete removes a profile by ID.
func (s *Store) Delete(id string) error {
	ps, err := s.Load()
	if err != nil {
		return err
	}
	for i, p := range ps {
		if p.ID == id {
			ps = append(ps[:i], ps[i+1:]...)
			return s.save(ps)
		}
	}
	return fmt.Errorf("profile %q not found", id)
}

// newID returns a random 8-char hex string.
func newID() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
