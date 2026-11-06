/*
Copyright © 2026 dos <debasmitbos22@gmail.com>
*/
package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gecko-ssg/gecko/config"
	"github.com/gecko-ssg/gecko/internal/markdown"
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
	RenderedHTML                                                               template.HTML
	Draft                                                                      bool
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
			Content:     page.RenderedHTML,
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
	renderer := markdown.New()
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
		result, err := renderer.Render(contents)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		page, err := sourcePageFromMetadata(result.Frontmatter)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		page.RenderedHTML = template.HTML(result.HTML)
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
	metadata, body, err := markdown.ParseFrontmatter([]byte(contents))
	if err != nil {
		return sourcePage{}, err
	}
	page, err := sourcePageFromMetadata(metadata)
	if err != nil {
		return sourcePage{}, err
	}
	page.Body = string(body)
	return page, nil
}

func sourcePageFromMetadata(metadata markdown.Frontmatter) (sourcePage, error) {
	if metadata.Title == "" || metadata.Template == "" {
		return sourcePage{}, fmt.Errorf("frontmatter requires title and template")
	}
	return sourcePage{
		Title:       metadata.Title,
		Description: metadata.Description,
		Template:    metadata.Template,
		Layout:      metadata.Layout,
		URL:         metadata.URL,
		Robots:      metadata.Robots,
		OGType:      metadata.OGType,
		Image:       metadata.Image,
		ImageAlt:    metadata.ImageAlt,
		Date:        metadata.Date,
		Tags:        metadata.Tags,
		Draft:       metadata.Draft,
	}, nil
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
