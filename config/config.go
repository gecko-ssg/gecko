package config

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

// Site describes the site-wide settings and portfolio content used while
// rendering a Gecko project.
type Site struct {
	Title       string       `yaml:"title"`
	Description string       `yaml:"description"`
	Language    string       `yaml:"language"`
	Author      string       `yaml:"author"`
	URL         string       `yaml:"url"`
	BaseURL     string       `yaml:"baseURL"`
	Logo        string       `yaml:"logo"`
	PostsDir    string       `yaml:"postsDir"`
	Year        int          `yaml:"year"`
	SEO         SEO          `yaml:"seo"`
	Theme       Theme        `yaml:"theme"`
	Profile     Profile      `yaml:"profile"`
	Projects    []Project    `yaml:"projects"`
	Experience  []Experience `yaml:"experience"`
}

// SEO contains the site's default search and social metadata.
type SEO struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Robots      string    `yaml:"robots"`
	Image       string    `yaml:"image"`
	ImageAlt    string    `yaml:"imageAlt"`
	OpenGraph   OpenGraph `yaml:"openGraph"`
	Twitter     Twitter   `yaml:"twitter"`
}

// OpenGraph contains site-wide Open Graph metadata.
type OpenGraph struct {
	Type     string `yaml:"type"`
	Locale   string `yaml:"locale"`
	SiteName string `yaml:"siteName"`
}

// Twitter contains site-wide X/Twitter card metadata.
type Twitter struct {
	Card string `yaml:"card"`
}

// Theme contains presentation settings exposed to templates.
type Theme struct {
	Color  string `yaml:"color"`
	Accent string `yaml:"accent"`
	Scheme string `yaml:"scheme"`
}

// Profile contains the portfolio owner information.
type Profile struct {
	Name     string       `yaml:"name"`
	Headline string       `yaml:"headline"`
	Bio      string       `yaml:"bio"`
	Image    string       `yaml:"image"`
	Email    string       `yaml:"email"`
	Socials  []SocialLink `yaml:"socials"`
}

// SocialLink is an ordered profile link displayed by the portfolio template.
type SocialLink struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Project is a project shown in the portfolio template.
type Project struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	URL          string   `yaml:"url"`
	Image        string   `yaml:"image"`
	Technologies []string `yaml:"technologies"`
}

// Experience is one role shown in the portfolio timeline.
type Experience struct {
	Company     string `yaml:"company"`
	Role        string `yaml:"role"`
	Period      string `yaml:"period"`
	Description string `yaml:"description"`
}

type file struct {
	Site       Site         `yaml:"site"`
	Profile    *Profile     `yaml:"profile"`
	Projects   []Project    `yaml:"projects"`
	Experience []Experience `yaml:"experience"`
}

// Load reads site.yaml and applies defaults used by the build pipeline.
func Load(path string) (Site, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Site{}, fmt.Errorf("read site config: %w", err)
	}

	var document file
	if err := yaml.Unmarshal(contents, &document); err != nil {
		return Site{}, fmt.Errorf("parse site config: %w", err)
	}

	site := document.Site
	if document.Profile != nil {
		site.Profile = *document.Profile
	}
	if document.Projects != nil {
		site.Projects = document.Projects
	}
	if document.Experience != nil {
		site.Experience = document.Experience
	}
	if site.Language == "" {
		site.Language = "en"
	}
	if site.Year == 0 {
		site.Year = time.Now().Year()
	}
	if site.SEO.Robots == "" {
		site.SEO.Robots = "index, follow"
	}
	if site.Theme.Color == "" {
		site.Theme.Color = "#ffffff"
	}
	if site.Title == "" {
		return Site{}, fmt.Errorf("site.yaml must define site.title")
	}
	return site, nil
}
