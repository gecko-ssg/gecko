/*
Copyright © 2026 dos <debasmitbos22@gmail.com>
*/
package cmd

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type siteConfig struct {
	Title, Description, Language, Author, URL, BaseURL, Logo, PostsDir string
	Year                                                               int
	SEO                                                                seoConfig
	Theme                                                              themeConfig
}
type seoConfig struct {
	Title, Description, Robots, Image, ImageAlt string
	OpenGraph                                   openGraphConfig
	Twitter                                     twitterConfig
}
type openGraphConfig struct{ Type, Locale, SiteName string }
type twitterConfig struct{ Card string }
type themeConfig struct{ Color string }

type pageData struct {
	Site                                                                siteConfig
	Title, Description, URL, Canonical, Robots, OGType, Image, ImageAlt string
	Date                                                                time.Time
	ReadingTime                                                         int
	Tags                                                                []string
	Content                                                             template.HTML
	Posts, RecentPosts                                                  []postSummary
	Projects                                                            []project
	Experience                                                          []experience
	Profile                                                             profile
}
type postSummary struct {
	Title, Description, URL string
	Date                    time.Time
}
type project struct {
	Title, Description, URL, Image string
	Technologies                   []string
}
type experience struct{ Role, Company, Period, Description string }
type profile struct {
	Name, Headline, Bio, Image, Email string
	Socials                           []social
}
type social struct{ Name, URL string }
type sourcePage struct {
	Title, Description, Template, Layout, URL, Robots, OGType, Image, ImageAlt string
	Date                                                                       time.Time
	Tags                                                                       []string
	Body, SourcePath                                                           string
}

var buildCmd = &cobra.Command{
	Use: "build", Short: "Build the site into dist",
	Long: "Render Markdown and Go HTML templates into static files in dist/ and copy assets and static files.",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := buildProject(); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Built site: dist/")
		return nil
	},
}

func init() { rootCmd.AddCommand(buildCmd) }

func buildProject() error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return buildSite(root)
}

func buildSite(root string) error {
	return buildSiteTo(root, filepath.Join(root, "dist"))
}

func buildSiteTo(root, dist string) error {
	config, err := readSiteConfig(filepath.Join(root, "site.yaml"))
	if err != nil {
		return err
	}
	pages, err := readContent(filepath.Join(root, "content"))
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dist); err != nil {
		return fmt.Errorf("clean dist: %w", err)
	}
	if err := os.MkdirAll(dist, 0o755); err != nil {
		return fmt.Errorf("create dist: %w", err)
	}
	if err := copyTree(filepath.Join(root, "assets"), filepath.Join(dist, "assets")); err != nil {
		return err
	}
	if err := copyTree(filepath.Join(root, "static"), dist); err != nil {
		return err
	}

	posts := make([]postSummary, 0)
	for _, page := range pages {
		if page.Template == "post" {
			posts = append(posts, postSummary{page.Title, page.Description, page.URL, page.Date})
		}
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].Date.After(posts[j].Date) })
	for _, page := range pages {
		data := pageData{Site: config, Title: page.Title, Description: page.Description, URL: page.URL, Robots: page.Robots, OGType: page.OGType, Image: page.Image, ImageAlt: page.ImageAlt, Date: page.Date, Tags: page.Tags, Content: template.HTML(renderMarkdown(page.Body)), Posts: posts, RecentPosts: posts}
		if err := renderPage(root, dist, page, data); err != nil {
			return err
		}
	}
	return nil
}

func readSiteConfig(path string) (siteConfig, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return siteConfig{}, fmt.Errorf("read site.yaml: %w", err)
	}
	config := siteConfig{Language: "en", Year: time.Now().Year(), SEO: seoConfig{Robots: "index, follow"}, Theme: themeConfig{Color: "#ffffff"}}
	section := ""
	for _, line := range strings.Split(string(contents), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, ":") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		parts := strings.SplitN(trimmed, ":", 2)
		key, rawValue := parts[0], strings.TrimSpace(parts[1])
		value := strings.Trim(rawValue, "\"'")
		if rawValue == "" && indent <= 2 {
			section = key
			continue
		}
		if indent == 0 {
			section = ""
		}
		switch section + "." + key {
		case "site.title":
			config.Title = value
		case "site.description":
			config.Description = value
		case "site.language":
			config.Language = value
		case "site.author":
			config.Author = value
		case "site.url":
			config.URL = value
		case "site.baseURL":
			config.BaseURL = value
		case "site.logo":
			config.Logo = value
		case "site.postsDir":
			config.PostsDir = value
		case "site.year":
			fmt.Sscanf(value, "%d", &config.Year)
		case "seo.title":
			config.SEO.Title = value
		case "seo.description":
			config.SEO.Description = value
		case "seo.robots":
			config.SEO.Robots = value
		case "seo.image":
			config.SEO.Image = value
		case "seo.imageAlt":
			config.SEO.ImageAlt = value
		case "theme.color":
			config.Theme.Color = value
		}
	}
	if config.Title == "" {
		return siteConfig{}, fmt.Errorf("site.yaml must define site.title")
	}
	return config, nil
}

func readContent(directory string) ([]sourcePage, error) {
	var pages []sourcePage
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		page, err := parsePage(string(contents))
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		page.SourcePath = path
		if page.URL == "" {
			page.URL = routeFor(path, directory)
		}
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read content: %w", err)
	}
	return pages, nil
}

func parsePage(contents string) (sourcePage, error) {
	if !strings.HasPrefix(contents, "---\n") {
		return sourcePage{}, fmt.Errorf("missing YAML front matter")
	}
	parts := strings.SplitN(contents[4:], "\n---\n", 2)
	if len(parts) != 2 {
		return sourcePage{}, fmt.Errorf("unterminated YAML front matter")
	}
	page := sourcePage{Body: parts[1]}
	for _, line := range strings.Split(parts[0], "\n") {
		pair := strings.SplitN(line, ":", 2)
		if len(pair) != 2 {
			continue
		}
		key, value := strings.TrimSpace(pair[0]), strings.Trim(strings.TrimSpace(pair[1]), "\"'")
		switch key {
		case "title":
			page.Title = value
		case "description":
			page.Description = value
		case "template":
			page.Template = value
		case "layout":
			page.Layout = value
		case "url":
			page.URL = value
		case "robots":
			page.Robots = value
		case "ogType":
			page.OGType = value
		case "image":
			page.Image = value
		case "imageAlt":
			page.ImageAlt = value
		case "date":
			page.Date, _ = time.Parse("2006-01-02", value)
		case "tags":
			page.Tags = parseList(value)
		}
	}
	if page.Title == "" || page.Template == "" {
		return sourcePage{}, fmt.Errorf("front matter requires title and template")
	}
	return page, nil
}

func parseList(value string) []string {
	value = strings.Trim(value, "[]")
	if value == "" {
		return nil
	}
	items := strings.Split(value, ",")
	for i := range items {
		items[i] = strings.Trim(strings.TrimSpace(items[i]), "\"'")
	}
	return items
}

func routeFor(path, contentRoot string) string {
	rel, _ := filepath.Rel(contentRoot, path)
	name := strings.TrimSuffix(filepath.ToSlash(rel), ".md")
	if name == "index" {
		return "/"
	}
	return "/" + name + "/"
}

func renderPage(root, dist string, page sourcePage, data pageData) error {
	base := filepath.Join(root, "layouts", "base.html")
	child := filepath.Join(root, "templates", page.Template+".html")
	if page.Template == "post" {
		child = filepath.Join(root, "layouts", "post.html")
	}
	tmpl, err := template.ParseFiles(base, child)
	if err != nil {
		return fmt.Errorf("parse templates for %s: %w", page.SourcePath, err)
	}
	var rendered bytes.Buffer
	if err := tmpl.ExecuteTemplate(&rendered, "base", data); err != nil {
		return fmt.Errorf("render %s: %w", page.SourcePath, err)
	}
	output := filepath.Join(dist, strings.TrimPrefix(page.URL, "/"), "index.html")
	if page.URL == "/" {
		output = filepath.Join(dist, "index.html")
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(output, rendered.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", output, err)
	}
	return nil
}

func copyTree(source, destination string) error {
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		output, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer output.Close()
		_, err = io.Copy(output, input)
		return err
	})
}

func renderMarkdown(markdown string) string {
	var output strings.Builder
	inList, inParagraph := false, false
	closeParagraph := func() {
		if inParagraph {
			output.WriteString("</p>\n")
			inParagraph = false
		}
	}
	for _, raw := range strings.Split(markdown, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case line == "":
			closeParagraph()
			if inList {
				output.WriteString("</ul>\n")
				inList = false
			}
		case strings.HasPrefix(line, "## "):
			closeParagraph()
			output.WriteString("<h2>" + html.EscapeString(line[3:]) + "</h2>\n")
		case strings.HasPrefix(line, "# "):
			closeParagraph()
			output.WriteString("<h1>" + html.EscapeString(line[2:]) + "</h1>\n")
		case strings.HasPrefix(line, "- "):
			closeParagraph()
			if !inList {
				output.WriteString("<ul>\n")
				inList = true
			}
			output.WriteString("<li>" + html.EscapeString(line[2:]) + "</li>\n")
		default:
			if !inParagraph {
				output.WriteString("<p>")
				inParagraph = true
			} else {
				output.WriteByte(' ')
			}
			output.WriteString(html.EscapeString(line))
		}
	}
	closeParagraph()
	if inList {
		output.WriteString("</ul>\n")
	}
	return output.String()
}
