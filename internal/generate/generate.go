package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/LordFarquaadtheCreator/cover-letter-writter/internal/profile"
)

// Input is the agent-facing generate_cover_letter tool input.
type Input struct {
	ProfileID string `json:"profileId" jsonschema:"required,ID of the profile to use (from create_profile or list_profiles). Mandatory."`
	Body      string `json:"body" jsonschema:"required,The cover letter body text. Plain text — the agent writes or drafts this."`
	OutputDir string `json:"outputDir,omitempty" jsonschema:"Override profile's outputDir. Defaults to ~/Downloads if neither set."`
	Filename  string `json:"filename,omitempty" jsonschema:"Override profile's filename. Defaults to <Name>NoSpacesCoverLetter.pdf."`
}

// Output is returned to the agent.
type Output struct {
	Message    string `json:"message"`
	OutputPath string `json:"outputPath"`
	Filename   string `json:"filename"`
}

// Run builds the PDF and writes it to disk. Core layout preserved from the
// original CLI: green sidebar, centered name header, two-column body with
// contact info on the right.
func Run(p profile.Profile, body, outputDir, filename string) (Output, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return Output{}, fmt.Errorf("body is empty")
	}

	pdf := fpdf.New("P", "mm", "Letter", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetAutoPageBreak(true, 25)
	pdf.AddPage()

	// green sidebar
	pdf.SetFillColor(0, 100, 40)
	pdf.Rect(0, 0, 5, 279.4, "F")

	// header
	pdf.SetY(20)
	pdf.SetFont("Times", "B", 28)
	pdf.CellFormat(0, 12, sanitize(tr, p.Name), "", 1, "C", false, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(20, pdf.GetY()+3, 215.9-20, pdf.GetY()+3)
	pdf.Ln(8)

	leftX := 20.0
	leftW := 120.0
	rightX := 145.0
	rightW := 45.0

	bodySecYCord := pdf.GetY()

	// salutation + date + body in left column
	pdf.SetXY(leftX, bodySecYCord)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.MultiCell(leftW, 5, sanitize(tr, "To Whom it May Concern,"), "", "L", false)

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetX(leftX)
	pdf.MultiCell(leftW, 5, sanitize(tr, time.Now().Format("01/02/2006")), "", "L", false)

	pdf.SetXY(leftX, pdf.GetY()+1)
	pdf.MultiCell(leftW, 5.5, sanitize(tr, body), "", "L", false)

	pdf.Ln(8)
	pdf.SetX(leftX)
	pdf.MultiCell(leftW, 5.5, sanitize(tr, "Best regards,\n"+p.Name), "", "L", false)

	// right column contact info
	pdf.SetXY(rightX, bodySecYCord)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(rightW, 4, "ADDRESS", "", 1, "L", false, 0, "")
	pdf.SetXY(rightX, pdf.GetY())
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Helvetica", "", 9)
	pdf.MultiCell(rightW, 4, sanitize(tr, p.Address), "", "L", false)
	pdf.Ln(1)

	pdf.SetX(rightX)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(rightW, 4, "EMAIL", "", 1, "L", false, 0, "")
	pdf.SetXY(rightX, pdf.GetY())
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(rightW, 4, sanitize(tr, p.Email), "", 1, "L", false, 0, "")
	pdf.Ln(1)

	pdf.SetX(rightX)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(rightW, 4, "PHONE", "", 1, "L", false, 0, "")
	pdf.SetXY(rightX, pdf.GetY())
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(rightW, 4, sanitize(tr, p.Phone), "", 1, "L", false, 0, "")

	dir := outputDir
	if dir == "" {
		dir = p.OutputDir
	}
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Output{}, fmt.Errorf("resolve output dir: %w", err)
		}
		dir = filepath.Join(home, "Downloads")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Output{}, fmt.Errorf("create output dir: %w", err)
	}

	fname := filename
	if fname == "" {
		fname = p.Filename
	}
	if fname == "" {
		fname = strings.ReplaceAll(p.Name, " ", "") + "CoverLetter.pdf"
	}

	outPath := filepath.Join(dir, fname)
	if err := pdf.OutputFileAndClose(outPath); err != nil {
		return Output{}, fmt.Errorf("write pdf: %w", err)
	}

	return Output{
		Message:    fmt.Sprintf("Cover letter saved to %s", outPath),
		OutputPath: outPath,
		Filename:   fname,
	}, nil
}

// sanitize swaps smart quotes/em dashes/ellipsis to ASCII before translation —
// fpdf's UnicodeTranslator cannot handle those glyphs.
func sanitize(tr func(string) string, s string) string {
	s = strings.ReplaceAll(s, "\u2014", "-")   // em dash
	s = strings.ReplaceAll(s, "\u2013", "-")   // en dash
	s = strings.ReplaceAll(s, "\u201C", `"`)   // left double quote
	s = strings.ReplaceAll(s, "\u201D", `"`)   // right double quote
	s = strings.ReplaceAll(s, "\u2018", "'")   // left single quote
	s = strings.ReplaceAll(s, "\u2019", "'")   // right single quote
	s = strings.ReplaceAll(s, "\u2026", "...") // ellipsis
	return tr(s)
}
