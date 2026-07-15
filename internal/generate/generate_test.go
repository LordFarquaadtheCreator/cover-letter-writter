package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LordFarquaadtheCreator/cover-letter-writter/internal/profile"
)

func TestRunEmptyBody(t *testing.T) {
	_, err := Run(profile.Profile{Name: "X", Email: "x@y.com"}, "", "   ", "", "")
	if err == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestRunWritesPDF(t *testing.T) {
	dir := t.TempDir()
	p := profile.Profile{Name: "Fahad Faruqi", Address: "NYC, NY", Email: "f@x.com", Phone: "(555) 123-4567"}
	out, err := Run(p, "", "I am writing to apply for the role.", dir, "Test.pdf")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.HasSuffix(out.OutputPath, "Test.pdf") {
		t.Fatalf("output path = %q", out.OutputPath)
	}
	info, err := os.Stat(out.OutputPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("PDF is empty")
	}
}

func TestRunDefaultsFilename(t *testing.T) {
	dir := t.TempDir()
	p := profile.Profile{Name: "Fahad Faruqi", Email: "f@x.com"}
	out, err := Run(p, "", "body text here", dir, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.Filename != "FahadFaruqiCoverLetter.pdf" {
		t.Fatalf("filename = %q", out.Filename)
	}
}

func TestRunCreatesOutputDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "deep")
	p := profile.Profile{Name: "X", Email: "x@y.com"}
	out, err := Run(p, "", "body", dir, "x.pdf")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out.OutputPath); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestRunOverridesTakePrecedence(t *testing.T) {
	dir := t.TempDir()
	p := profile.Profile{Name: "X", Email: "x@y.com", OutputDir: "/should/not/use", Filename: "bad.pdf"}
	out, err := Run(p, "", "body", dir, "good.pdf")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.HasSuffix(out.OutputPath, "good.pdf") || strings.Contains(out.OutputPath, "should") {
		t.Fatalf("override not applied: %q", out.OutputPath)
	}
}

func TestRunSanitizesSmartQuotes(t *testing.T) {
	dir := t.TempDir()
	p := profile.Profile{Name: "Test", Email: "t@x.com"}
	body := "I\u2019m excited \u2014 this is \u201cgreat\u201d\u2026"
	out, err := Run(p, "", body, dir, "sanitize.pdf")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out.OutputPath); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestRunDefaultsSalutation(t *testing.T) {
	dir := t.TempDir()
	p := profile.Profile{Name: "X", Email: "x@y.com"}
	if _, err := Run(p, "", "body", dir, "default.pdf"); err != nil {
		t.Fatalf("Run with empty to: %v", err)
	}
}

func TestRunCustomSalutation(t *testing.T) {
	dir := t.TempDir()
	p := profile.Profile{Name: "X", Email: "x@y.com"}
	if _, err := Run(p, "Dear Jane Doe,", "body", dir, "custom.pdf"); err != nil {
		t.Fatalf("Run with custom to: %v", err)
	}
}
