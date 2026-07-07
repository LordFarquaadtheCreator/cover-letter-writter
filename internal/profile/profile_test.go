package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func tmpStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	return NewStore(dir)
}

func TestLoadMissingFile(t *testing.T) {
	s := tmpStore(t)
	ps, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ps) != 0 {
		t.Fatalf("expected empty, got %d", len(ps))
	}
}

func TestCreateRequiresNameAndEmail(t *testing.T) {
	s := tmpStore(t)
	if _, err := s.Create(Profile{Email: "a@b.c"}); err == nil {
		t.Fatal("expected error for missing name")
	}
	if _, err := s.Create(Profile{Name: "X"}); err == nil {
		t.Fatal("expected error for missing email")
	}
}

func TestCreateAndGet(t *testing.T) {
	s := tmpStore(t)
	p, err := s.Create(Profile{Name: "Fahad", Email: "f@x.com", Phone: "555"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.ID == "" {
		t.Fatal("empty ID")
	}
	got, err := s.Get(p.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Fahad" {
		t.Fatalf("name = %q", got.Name)
	}
}

func TestGetNotFound(t *testing.T) {
	s := tmpStore(t)
	if _, err := s.Get("nope"); err == nil {
		t.Fatal("expected not found error")
	}
}

func TestListSortedByCreated(t *testing.T) {
	s := tmpStore(t)
	p1, _ := s.Create(Profile{Name: "A", Email: "a@x.com"})
	p2, _ := s.Create(Profile{Name: "B", Email: "b@x.com"})
	ps, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ps) != 2 || ps[0].ID != p1.ID || ps[1].ID != p2.ID {
		t.Fatalf("order wrong: %v %v", ps[0].ID, ps[1].ID)
	}
}

func TestUpdateAppliesNonEmptyFields(t *testing.T) {
	s := tmpStore(t)
	p, _ := s.Create(Profile{Name: "Old", Email: "o@x.com", Phone: "111"})
	updated, err := s.Update(p.ID, Profile{Name: "New", Phone: "222"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "New" || updated.Phone != "222" {
		t.Fatalf("fields not applied: %+v", updated)
	}
	if updated.Email != "o@x.com" {
		t.Fatalf("email should be unchanged: %q", updated.Email)
	}
}

func TestUpdateNotFound(t *testing.T) {
	s := tmpStore(t)
	if _, err := s.Update("nope", Profile{Name: "X"}); err == nil {
		t.Fatal("expected not found error")
	}
}

func TestDelete(t *testing.T) {
	s := tmpStore(t)
	p, _ := s.Create(Profile{Name: "A", Email: "a@x.com"})
	if err := s.Delete(p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get(p.ID); err == nil {
		t.Fatal("profile still exists")
	}
}

func TestDeleteNotFound(t *testing.T) {
	s := tmpStore(t)
	if err := s.Delete("nope"); err == nil {
		t.Fatal("expected not found error")
	}
}

func TestPersistsAcrossStoreInstances(t *testing.T) {
	dir := t.TempDir()
	s1 := NewStore(dir)
	p, _ := s1.Create(Profile{Name: "Fahad", Email: "f@x.com"})
	s2 := NewStore(dir)
	got, err := s2.Get(p.ID)
	if err != nil {
		t.Fatalf("Get from new store: %v", err)
	}
	if got.Name != "Fahad" {
		t.Fatalf("name = %q", got.Name)
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "profiles.json"), []byte("{bad"), 0644)
	s := NewStore(dir)
	if _, err := s.Load(); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "profiles.json"), []byte(""), 0644)
	s := NewStore(dir)
	ps, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ps) != 0 {
		t.Fatalf("expected empty, got %d", len(ps))
	}
}
