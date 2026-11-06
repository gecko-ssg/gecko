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

	"github.com/gecko-ssg/gecko/config"
	"github.com/spf13/cobra"
)

type siteConfig = config.Site

type pageData struct {
	Site                                                                siteConfig
	Title, Description, URL, Canonical, Robots, OGType, Image, ImageAlt string
	Date                                                                time.Time
	ReadingTime                                                         int
	Tags                                                                []string
	Content                                                             template.HTML
	Posts, RecentPosts                                                  []postSummary
	Projects                                                            []config.Project
	Experience                                                          []config.Experience
	Profile                                                             config.Profile
}
type postSummary struct {
	Title, Description, URL string
	Date                    time.Time
}
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
		data := pageData{
			Site:        config,
			Title:       page.Title,
			Description: page.Description,
			URL:         page.URL,
			Robots:      page.Robots,
			OGType:      page.OGType,
			Image:       page.Image,
			ImageAlt:    page.ImageAlt,
			Date:        page.Date,
			Tags:        page.Tags,
			Content:     template.HTML(renderMarkdown(page.Body)),
			Posts:       posts,
			RecentPosts: posts,
			Profile:     config.Profile,
			Projects:    config.Projects,
			Experience:  config.Experience,
		}
		if err := renderPage(root, dist, page, data); err != nil {
			return err
		}
	}
	return nil
}

func readSiteConfig(path string) (siteConfig, error) {
	return config.Load(path)
}

func readContent(directory string) ([]sourcePage, error) {
	var pages []sourcePage
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		ext := filepath.Ext(path)
		if entry.IsDir() || (ext != ".md" && ext != ".gmd") {
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
	name := strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(path))
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
		case strings.HasPrefix(line, "# "):
			closeParagraph()
			output.WriteString("<h1>" + html.EscapeString(line[3:]) + "</h1>\n")
		case strings.HasPrefix(line, "## "):
			closeParagraph()
			output.WriteString("<h2>" + html.EscapeString(line[3:]) + "</h2>\n")
		case strings.HasPrefix(line, "### "):
			closeParagraph()
			output.WriteString("<h3>" + html.EscapeString(line[3:]) + "</h3>\n")
		case strings.HasPrefix(line, "#### "):
			closeParagraph()
			output.WriteString("<h4>" + html.EscapeString(line[3:]) + "</h4>\n")
		case strings.HasPrefix(line, "##### "):
			closeParagraph()
			output.WriteString("<h5>" + html.EscapeString(line[3:]) + "</h5>\n")
		case strings.HasPrefix(line, "###### "):
			closeParagraph()
			output.WriteString("<h6>" + html.EscapeString(line[2:]) + "</h6>\n")
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
