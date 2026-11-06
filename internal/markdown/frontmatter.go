package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

// Frontmatter contains metadata shared by Gecko's content pages.
type Frontmatter struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Date        time.Time `yaml:"date"`
	Draft       bool      `yaml:"draft"`
	Template    string    `yaml:"template"`
	Layout      string    `yaml:"layout"`
	URL         string    `yaml:"url"`
	Robots      string    `yaml:"robots"`
	OGType      string    `yaml:"ogType"`
	Image       string    `yaml:"image"`
	ImageAlt    string    `yaml:"imageAlt"`
	Tags        []string  `yaml:"tags"`
}

// ParseFrontmatter separates YAML metadata from the Markdown body.
func ParseFrontmatter(source []byte) (Frontmatter, []byte, error) {
	source = bytes.TrimPrefix(source, []byte("\xef\xbb\xbf"))
	lines := bytes.Split(source, []byte("\n"))
	if len(lines) == 0 || string(trimCR(lines[0])) != "---" {
		return Frontmatter{}, nil, fmt.Errorf("missing YAML frontmatter opening delimiter")
	}

	closing := -1
	for i := 1; i < len(lines); i++ {
		if string(trimCR(lines[i])) == "---" {
			closing = i
			break
		}
	}
	if closing < 0 {
		return Frontmatter{}, nil, fmt.Errorf("missing YAML frontmatter closing delimiter")
	}

	var metadata Frontmatter
	if err := yaml.Unmarshal(bytes.Join(lines[1:closing], []byte("\n")), &metadata); err != nil {
		return Frontmatter{}, nil, fmt.Errorf("parse YAML frontmatter: %w", err)
	}
	body := bytes.Join(lines[closing+1:], []byte("\n"))
	return metadata, body, nil
}

func trimCR(line []byte) []byte {
	return []byte(strings.TrimSuffix(string(line), "\r"))
}
