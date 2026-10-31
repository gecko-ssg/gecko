/*
Copyright © 2026 dos <debasmitbos22@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var contentSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var newCmd = &cobra.Command{Use: "new", Short: "Create content", Args: cobra.NoArgs}
var newPageCmd = &cobra.Command{
	Use: "page <title>", Short: "Create a Markdown page", Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createContent(cmd, strings.Join(args, " "), false)
	},
}
var newPostCmd = &cobra.Command{
	Use: "post <title>", Short: "Create a Markdown post", Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createContent(cmd, strings.Join(args, " "), true)
	},
}

func init() {
	newCmd.AddCommand(newPageCmd, newPostCmd)
	rootCmd.AddCommand(newCmd)
}

func createContent(cmd *cobra.Command, title string, post bool) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get project directory: %w", err)
	}
	slug := contentSlug(title)
	if !contentSlugPattern.MatchString(slug) {
		return fmt.Errorf("title %q does not produce a valid content filename", title)
	}
	directory := filepath.Join(root, "content")
	templateName, layout := "page", "base"
	if post {
		config, err := readSiteConfig(filepath.Join(root, "site.yaml"))
		if err != nil {
			return fmt.Errorf("load post directory from site.yaml: %w", err)
		}
		postsDir := config.PostsDir
		if postsDir == "" {
			postsDir = filepath.Join("content", "posts")
		}
		if filepath.IsAbs(postsDir) {
			return fmt.Errorf("site.postsDir must be relative to the project directory")
		}
		directory = filepath.Join(root, filepath.Clean(postsDir))
		if !isWithin(directory, root) || directory == root {
			return fmt.Errorf("site.postsDir must stay inside the project directory")
		}
		templateName, layout = "post", "post"
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create content directory: %w", err)
	}
	filename := filepath.Join(directory, slug+".md")
	if _, err := os.Lstat(filename); err == nil {
		return fmt.Errorf("content already exists: %s", filename)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check content path: %w", err)
	}
	var contents strings.Builder
	fmt.Fprintf(&contents, "---\ntitle: %s\n", strconv.Quote(title))
	if post {
		fmt.Fprintf(&contents, "date: %s\nurl: %q\n", time.Now().Format("2006-01-02"), "/blog/"+slug+"/")
	}
	fmt.Fprintf(&contents, "template: %q\nlayout: %q\n---\n\n# %s\n\n", templateName, layout, title)
	if err := os.WriteFile(filename, []byte(contents.String()), 0o644); err != nil {
		return fmt.Errorf("write content: %w", err)
	}
	kind := "page"
	if post {
		kind = "post"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created %s: %s\n", kind, filepath.ToSlash(strings.TrimPrefix(filename, root+string(filepath.Separator))))
	return nil
}

func contentSlug(title string) string {
	var result strings.Builder
	dash := false
	for _, char := range strings.ToLower(strings.TrimSpace(title)) {
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			result.WriteRune(char)
			dash = false
		} else if !dash && result.Len() > 0 {
			result.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(result.String(), "-")
}
