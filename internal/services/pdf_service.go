package services

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/phpdave11/gofpdf"
)

// PDFService handles PDF generation for song content
type PDFService struct{}

// NewPDFService creates a new PDF service instance
func NewPDFService() *PDFService {
	return &PDFService{}
}

// SongContentPDFRequest represents the request for PDF generation
type SongContentPDFRequest struct {
	SongTitle string `json:"song_title"`
	Artist    string `json:"artist"`
	Key       string `json:"key"`
	Tempo     *int   `json:"tempo"`
	Content   string `json:"content"`
}

// GenerateSongPDF generates a PDF from song content
func (s *PDFService) GenerateSongPDF(req *SongContentPDFRequest) ([]byte, error) {
	// Create a new PDF document with UTF-8 support
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Register DejaVu fonts for UTF-8 support
	pdf.AddUTF8Font("DejaVu", "", "fonts/DejaVuSans.ttf")
	pdf.AddUTF8Font("DejaVu", "B", "fonts/DejaVuSans-Bold.ttf")
	pdf.AddUTF8Font("DejaVu", "I", "fonts/DejaVuSans-Oblique.ttf")
	pdf.AddUTF8Font("DejaVu", "BI", "fonts/DejaVuSans-BoldOblique.ttf")

	// Set document metadata
	pdf.SetAuthor("Setlist Manager", false)
	pdf.SetCreator("Setlist Manager", false)
	pdf.SetTitle(fmt.Sprintf("%s - %s", req.SongTitle, req.Artist), false)

	pdf.AddPage()

	// Set margins
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// Title section with UTF-8 support
	pdf.SetFont("DejaVu", "B", 18)
	title := req.SongTitle
	if req.Artist != "" {
		title = fmt.Sprintf("%s - %s", req.SongTitle, req.Artist)
	}
	// Handle UTF-8 characters properly
	pdf.Cell(0, 10, title)
	pdf.Ln(15)

	// Song info section
	pdf.SetFont("DejaVu", "", 12)
	infoLine := ""
	if req.Key != "" {
		infoLine = fmt.Sprintf("Key: %s", req.Key)
	}
	if req.Tempo != nil {
		if infoLine != "" {
			infoLine += " | "
		}
		infoLine += fmt.Sprintf("Tempo: %d BPM", *req.Tempo)
	}
	if infoLine != "" {
		pdf.Cell(0, 8, infoLine)
		pdf.Ln(15)
	}

	// Content section
	pdf.SetFont("DejaVu", "", 11)

	// Split content into lines and process markdown formatting
	lines := strings.Split(req.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			pdf.Ln(5)
			continue
		}

		// Handle headers
		if strings.HasPrefix(line, "# ") {
			pdf.SetFont("DejaVu", "B", 16)
			pdf.Cell(0, 8, strings.TrimPrefix(line, "# "))
			pdf.Ln(8)
			pdf.SetFont("DejaVu", "", 11)
			continue
		}

		if strings.HasPrefix(line, "## ") {
			pdf.SetFont("DejaVu", "B", 14)
			pdf.Cell(0, 7, strings.TrimPrefix(line, "## "))
			pdf.Ln(7)
			pdf.SetFont("DejaVu", "", 11)
			continue
		}

		if strings.HasPrefix(line, "### ") {
			pdf.SetFont("DejaVu", "B", 12)
			pdf.Cell(0, 6, strings.TrimPrefix(line, "### "))
			pdf.Ln(6)
			pdf.SetFont("DejaVu", "", 11)
			continue
		}

		// Handle bold text
		if strings.Contains(line, "**") {
			parts := strings.Split(line, "**")
			for i, part := range parts {
				if i%2 == 1 { // Bold text
					pdf.SetFont("DejaVu", "B", 11)
					pdf.Write(5, part)
					pdf.SetFont("DejaVu", "", 11)
				} else { // Regular text
					pdf.Write(5, part)
				}
			}
			pdf.Ln(5)
			continue
		}

		// Handle italic text
		if strings.Contains(line, "*") && !strings.Contains(line, "**") {
			parts := strings.Split(line, "*")
			for i, part := range parts {
				if i%2 == 1 && i < len(parts)-1 { // Italic text (but not the last part if it ends with *)
					pdf.SetFont("DejaVu", "I", 11)
					pdf.Write(5, part)
					pdf.SetFont("DejaVu", "", 11)
				} else { // Regular text
					pdf.Write(5, part)
				}
			}
			pdf.Ln(5)
			continue
		}

		// Handle bullet points
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			pdf.Cell(5, 5, "•")
			pdf.Write(5, " "+strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
			pdf.Ln(5)
			continue
		}

		// Handle numbered lists
		if strings.Contains(line, ". ") && len(line) > 2 {
			// Simple check for numbered lists
			parts := strings.SplitN(line, ". ", 2)
			if len(parts) == 2 {
				pdf.Write(5, parts[0]+". "+parts[1])
				pdf.Ln(5)
				continue
			}
		}

		// Handle code blocks (simple inline code)
		if strings.Contains(line, "`") {
			parts := strings.Split(line, "`")
			for i, part := range parts {
				if i%2 == 1 { // Code text
					pdf.SetFont("DejaVu", "", 10) // Use DejaVu for code too
					pdf.Write(5, part)
					pdf.SetFont("DejaVu", "", 11)
				} else { // Regular text
					pdf.Write(5, part)
				}
			}
			pdf.Ln(5)
			continue
		}

		// Regular text
		pdf.Write(5, line)
		pdf.Ln(5)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
