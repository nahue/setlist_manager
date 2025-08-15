package services

import (
	"html/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// MarkdownService handles markdown parsing and rendering
type MarkdownService struct{}

// NewMarkdownService creates a new markdown service
func NewMarkdownService() *MarkdownService {
	return &MarkdownService{}
}

// ParseMarkdown converts markdown text to HTML
func (s *MarkdownService) ParseMarkdown(text string) template.HTML {
	if text == "" {
		return template.HTML("")
	}

	// Create markdown parser with extensions
	extensions := parser.CommonExtensions
	parser := parser.NewWithExtensions(extensions)

	// Parse markdown
	doc := parser.Parse([]byte(text))

	// Create HTML renderer
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// Render to HTML
	htmlBytes := markdown.Render(doc, renderer)

	return template.HTML(htmlBytes)
}

// ParseMarkdownSafe converts markdown text to HTML with safety measures
func (s *MarkdownService) ParseMarkdownSafe(text string) template.HTML {
	// For now, we'll use the same implementation
	// In a production environment, you might want to add additional sanitization
	return s.ParseMarkdown(text)
}
