// Package markdown parses Markdown content and renders it as safe HTML.
package markdown

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	htmlrenderer "github.com/yuin/goldmark/renderer/html"
)

// Renderer converts Markdown documents into HTML and extracts frontmatter.
type Renderer struct {
	parser goldmark.Markdown
}

// RenderResult is the rendered HTML and its document metadata.
type RenderResult struct {
	HTML        []byte
	Frontmatter Frontmatter
}

// New creates a Markdown renderer with common extensions and heading IDs.
// Raw HTML remains disabled so untrusted Markdown cannot inject markup.
func New() *Renderer {
	return &Renderer{parser: goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.DefinitionList,
			extension.Typographer,
			highlighting.NewHighlighting(highlighting.WithStyle("github")),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(htmlrenderer.WithXHTML()),
	)}
}

// Render parses frontmatter, converts the Markdown body, and returns both.
func (r *Renderer) Render(source []byte) (*RenderResult, error) {
	metadata, body, err := ParseFrontmatter(source)
	if err != nil {
		return nil, fmt.Errorf("frontmatter: %w", err)
	}

	var output bytes.Buffer
	if err := r.parser.Convert(body, &output); err != nil {
		return nil, fmt.Errorf("render Markdown: %w", err)
	}

	return &RenderResult{HTML: output.Bytes(), Frontmatter: metadata}, nil
}
