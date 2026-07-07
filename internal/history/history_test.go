package history

import (
	"os"
	"path/filepath"
	"testing"
)

func tmpStore(t *testing.T) *Store {
	t.Helper()
	return NewStore(t.TempDir())
}

func TestLoadMissingFile(t *testing.T) {
	s := tmpStore(t)
	es, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(es) != 0 {
		t.Fatalf("expected empty, got %d", len(es))
	}
}

func TestAddRequiresProfileID(t *testing.T) {
	s := tmpStore(t)
	if _, err := s.Add(Entry{Body: "x"}); err == nil {
		t.Fatal("expected error for missing profileId")
	}
}

func TestAddAndList(t *testing.T) {
	s := tmpStore(t)
	e, err := s.Add(Entry{ProfileID: "p1", ProfileName: "Fahad", Body: "hello", OutputPath: "/tmp/x.pdf"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if e.ID == "" {
		t.Fatal("empty ID")
	}
	es, err := s.List("")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(es) != 1 || es[0].ID != e.ID {
		t.Fatalf("list mismatch: %+v", es)
	}
}

func TestListFiltersByProfile(t *testing.T) {
	s := tmpStore(t)
	s.Add(Entry{ProfileID: "p1", Body: "a"})
	s.Add(Entry{ProfileID: "p2", Body: "b"})
	es, err := s.List("p1")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(es) != 1 || es[0].ProfileID != "p1" {
		t.Fatalf("filter mismatch: %+v", es)
	}
}

func TestListNewestFirst(t *testing.T) {
	s := tmpStore(t)
	e1, _ := s.Add(Entry{ProfileID: "p1", Body: "first"})
	e2, _ := s.Add(Entry{ProfileID: "p1", Body: "second"})
	es, _ := s.List("")
	if es[0].ID != e2.ID || es[1].ID != e1.ID {
		t.Fatalf("order wrong: %+v", es)
	}
}

func TestPersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	s1 := NewStore(dir)
	e, _ := s1.Add(Entry{ProfileID: "p1", Body: "x"})
	s2 := NewStore(dir)
	es, err := s2.List("")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(es) != 1 || es[0].ID != e.ID {
		t.Fatalf("persist mismatch: %+v", es)
	}
}

func TestLoadMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "history.json"), []byte("{bad"), 0644)
	s := NewStore(dir)
	if _, err := s.Load(); err == nil {
		t.Fatal("expected parse error")
	}
}
