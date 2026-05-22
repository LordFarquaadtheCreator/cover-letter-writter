package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

type User struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	OutputDir string `json:"outputDir"`
	Filename  string `json:"filename"`
}

func main() {
	userData, err := os.ReadFile("user.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read user.json: %v\n", err)
		os.Exit(1)
	}
	var user User
	if err := json.Unmarshal(userData, &user); err != nil {
		fmt.Fprintf(os.Stderr, "parse user.json: %v\n", err)
		os.Exit(1)
	}

	bodyBytes, err := exec.Command("pbpaste").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "pbpaste: %v\n", err)
		os.Exit(1)
	}
	body := strings.TrimSpace(string(bodyBytes))
	if body == "" {
		fmt.Fprintln(os.Stderr, "clipboard empty")
		os.Exit(1)
	}

	if err := generatePDF(user, body); err != nil {
		fmt.Fprintf(os.Stderr, "generate pdf: %v\n", err)
		os.Exit(1)
	}
}

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

func generatePDF(user User, body string) error {
	pdf := fpdf.New("P", "mm", "Letter", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetAutoPageBreak(true, 25)
	pdf.AddPage()

	// green sidebar
	green := 100
	pdf.SetFillColor(0, green, 40)
	pdf.Rect(0, 0, 5, 279.4, "F")

	// header
	pdf.SetY(20)
	pdf.SetFont("Times", "B", 28)
	pdf.CellFormat(0, 12, sanitize(tr, user.Name), "", 1, "C", false, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(20, pdf.GetY()+3, 215.9-20, pdf.GetY()+3)
	pdf.Ln(8)

	// columns
	leftX := 20.0
	leftW := 120.0
	rightX := 145.0
	rightW := 45.0

	// y cord of body section
	bodySecYCord := pdf.GetY()

	// salutation, date, cover letter in left column
	pdf.SetXY(leftX, bodySecYCord)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.MultiCell(leftW, 5, sanitize(tr, "To Whom it May Concern,"), "", "L", false)

	pdf.SetFont("Helvetica", "", 10)
	dateStr := time.Now().Format("01/02/2006")
	pdf.SetX(leftX)
	pdf.MultiCell(leftW, 5, sanitize(tr, dateStr), "", "L", false)

	// body
	pdf.SetXY(leftX, pdf.GetY()+1)
	pdf.MultiCell(leftW, 5.5, sanitize(tr, body), "", "L", false)

	// closing
	pdf.Ln(8)
	pdf.SetX(leftX)
	pdf.MultiCell(leftW, 5.5, sanitize(tr, "Best regards,\n"+user.Name), "", "L", false)

	// right column contact info
	pdf.SetXY(rightX, bodySecYCord)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(rightW, 4, "ADDRESS", "", 1, "L", false, 0, "")
	pdf.SetXY(rightX, pdf.GetY())
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Helvetica", "", 9)
	pdf.MultiCell(rightW, 4, sanitize(tr, user.Address), "", "L", false)
	pdf.Ln(1)

	pdf.SetX(rightX)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(rightW, 4, "EMAIL", "", 1, "L", false, 0, "")
	pdf.SetXY(rightX, pdf.GetY())
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(rightW, 4, sanitize(tr, user.Email), "", 1, "L", false, 0, "")
	pdf.Ln(1)

	pdf.SetX(rightX)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(rightW, 4, "PHONE", "", 1, "L", false, 0, "")
	pdf.SetXY(rightX, pdf.GetY())
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(rightW, 4, sanitize(tr, user.Phone), "", 1, "L", false, 0, "")

	outDir := user.OutputDir
	if outDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		outDir = filepath.Join(home, "Downloads")
	}
	fname := user.Filename
	if fname == "" {
		fname = strings.ReplaceAll(user.Name, " ", "") + "CoverLetter.pdf"
	}
	return pdf.OutputFileAndClose(filepath.Join(outDir, fname))
}
