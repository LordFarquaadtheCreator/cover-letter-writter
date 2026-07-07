package mcpserver

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/LordFarquaadtheCreator/cover-letter-writter/internal/generate"
	"github.com/LordFarquaadtheCreator/cover-letter-writter/internal/history"
	"github.com/LordFarquaadtheCreator/cover-letter-writter/internal/profile"
)

func tmpDeps(t *testing.T) deps {
	t.Helper()
	dir := t.TempDir()
	return deps{
		Profiles: profile.NewStore(dir),
		History:  history.NewStore(dir),
	}
}

func TestHandleCreateProfile(t *testing.T) {
	d := tmpDeps(t)
	res, out, err := handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{
		Name: "Fahad", Email: "f@x.com", Phone: "555",
	}, d)
	if err != nil {
		t.Fatalf("handleCreateProfile: %v", err)
	}
	if out.Profile.ID == "" {
		t.Fatal("empty ID")
	}
	if len(res.Content) == 0 {
		t.Fatal("no content")
	}
}

func TestHandleCreateProfileMissingName(t *testing.T) {
	d := tmpDeps(t)
	_, _, err := handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{
		Email: "f@x.com",
	}, d)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleGetProfile(t *testing.T) {
	d := tmpDeps(t)
	_, out, err := handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{
		Name: "Fahad", Email: "f@x.com",
	}, d)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, got, err := handleGetProfile(context.Background(), &mcp.CallToolRequest{}, GetProfileInput{ID: out.Profile.ID}, d)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Profile.Name != "Fahad" {
		t.Fatalf("name = %q", got.Profile.Name)
	}
}

func TestHandleGetProfileNotFound(t *testing.T) {
	d := tmpDeps(t)
	_, _, err := handleGetProfile(context.Background(), &mcp.CallToolRequest{}, GetProfileInput{ID: "nope"}, d)
	if err == nil {
		t.Fatal("expected not found")
	}
}

func TestHandleListProfiles(t *testing.T) {
	d := tmpDeps(t)
	handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{Name: "A", Email: "a@x.com"}, d)
	handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{Name: "B", Email: "b@x.com"}, d)
	_, out, err := handleListProfiles(context.Background(), &mcp.CallToolRequest{}, d)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(out.Profiles) != 2 {
		t.Fatalf("expected 2, got %d", len(out.Profiles))
	}
}

func TestHandleUpdateProfile(t *testing.T) {
	d := tmpDeps(t)
	_, created, _ := handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{Name: "Old", Email: "o@x.com"}, d)
	_, out, err := handleUpdateProfile(context.Background(), &mcp.CallToolRequest{}, UpdateProfileInput{ID: created.Profile.ID, Name: "New"}, d)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if out.Profile.Name != "New" {
		t.Fatalf("name = %q", out.Profile.Name)
	}
	if out.Profile.Email != "o@x.com" {
		t.Fatalf("email changed: %q", out.Profile.Email)
	}
}

func TestHandleDeleteProfile(t *testing.T) {
	d := tmpDeps(t)
	_, created, _ := handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{Name: "X", Email: "x@x.com"}, d)
	_, _, err := handleDeleteProfile(context.Background(), &mcp.CallToolRequest{}, DeleteProfileInput{ID: created.Profile.ID}, d)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, _, err = handleGetProfile(context.Background(), &mcp.CallToolRequest{}, GetProfileInput{ID: created.Profile.ID}, d)
	if err == nil {
		t.Fatal("profile still exists")
	}
}

func TestHandleGenerate(t *testing.T) {
	d := tmpDeps(t)
	_, created, _ := handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{
		Name: "Fahad", Email: "f@x.com", Phone: "555", Address: "NYC",
		OutputDir: t.TempDir(),
	}, d)
	_, out, err := handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{
		ProfileID: created.Profile.ID,
		Body:      "I am writing to apply.",
	}, d)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !strings.HasSuffix(out.OutputPath, ".pdf") {
		t.Fatalf("output path = %q", out.OutputPath)
	}
}

func TestHandleGenerateMissingProfileID(t *testing.T) {
	d := tmpDeps(t)
	_, _, err := handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{Body: "x"}, d)
	if err == nil {
		t.Fatal("expected error for missing profileId")
	}
}

func TestHandleGenerateEmptyBody(t *testing.T) {
	d := tmpDeps(t)
	_, created, _ := handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{Name: "X", Email: "x@x.com"}, d)
	_, _, err := handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{ProfileID: created.Profile.ID}, d)
	if err == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestHandleGenerateRecordsHistory(t *testing.T) {
	d := tmpDeps(t)
	_, created, _ := handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{
		Name: "Fahad", Email: "f@x.com", OutputDir: t.TempDir(),
	}, d)
	handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{
		ProfileID: created.Profile.ID, Body: "test body",
	}, d)
	_, hist, err := handleListHistory(context.Background(), &mcp.CallToolRequest{}, ListHistoryInput{ProfileID: created.Profile.ID}, d)
	if err != nil {
		t.Fatalf("listHistory: %v", err)
	}
	if len(hist.Entries) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(hist.Entries))
	}
	if hist.Entries[0].Body != "test body" {
		t.Fatalf("body = %q", hist.Entries[0].Body)
	}
}

func TestHandleListHistoryFilter(t *testing.T) {
	d := tmpDeps(t)
	_, p1, _ := handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{Name: "A", Email: "a@x.com", OutputDir: t.TempDir()}, d)
	_, p2, _ := handleCreateProfile(context.Background(), &mcp.CallToolRequest{}, CreateProfileInput{Name: "B", Email: "b@x.com", OutputDir: t.TempDir()}, d)
	handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{ProfileID: p1.Profile.ID, Body: "a"}, d)
	handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{ProfileID: p2.Profile.ID, Body: "b"}, d)
	_, out, err := handleListHistory(context.Background(), &mcp.CallToolRequest{}, ListHistoryInput{ProfileID: p1.Profile.ID}, d)
	if err != nil {
		t.Fatalf("listHistory: %v", err)
	}
	if len(out.Entries) != 1 || out.Entries[0].ProfileID != p1.Profile.ID {
		t.Fatalf("filter mismatch: %+v", out.Entries)
	}
}
