/*
Copyright © 2026 dos <debasmitbos22@gmail.com>
*/
package cmd

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"
)

const defaultProjectName = "my-app"

var availableTemplates = []string{"portfolio"}

var projectNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

//go:embed all:templates/**
var starterTemplates embed.FS

var initGit bool
var initTemplate string

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Create a new Gecko site",
	Long:  "Create a new Gecko site from an embedded starter template. No network connection is required.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := defaultProjectName
		if len(args) == 1 {
			name = args[0]
		}
		return createProject(cmd, name, initTemplate, initGit)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initGit, "git", false, "Initialize a Git repository")
	initCmd.Flags().StringVar(&initTemplate, "template", "portfolio", "Choose a starter template")
}

func createProject(cmd *cobra.Command, name, templateName string, initializeGit bool) error {
	if !projectNamePattern.MatchString(name) {
		return fmt.Errorf("invalid project name %q: use letters, numbers, hyphens, and underscores only", name)
	}
	if !isAvailableTemplate(templateName) {
		return fmt.Errorf("unknown template %q; available templates: %s", templateName, strings.Join(availableTemplates, ", "))
	}
	source := filepath.ToSlash(filepath.Join("templates", templateName))
	if _, err := fs.Stat(starterTemplates, source); err != nil {
		return fmt.Errorf("template %q is unavailable: %w", templateName, err)
	}

	target, err := filepath.Abs(name)
	if err != nil {
		return fmt.Errorf("resolve project directory: %w", err)
	}
	if _, err := os.Lstat(target); err == nil {
		return fmt.Errorf("refusing to overwrite existing project: %s", target)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check project directory: %w", err)
	}
	if initializeGit {
		if _, err := exec.LookPath("git"); err != nil {
			return fmt.Errorf("git initialization requested but git is not installed")
		}
	}
	configValues, err := promptSiteConfig(cmd, name)
	if err != nil {
		return err
	}

	if err := os.Mkdir(target, 0o755); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}
	cleanup := func(cause error) error {
		if cleanupErr := os.RemoveAll(target); cleanupErr != nil {
			return errors.Join(cause, fmt.Errorf("clean up project directory: %w", cleanupErr))
		}
		return cause
	}
	if err := extractTemplateWithConfig(target, name, templateName, configValues); err != nil {
		return cleanup(err)
	}
	if initializeGit {
		git := exec.Command("git", "init")
		git.Dir = target
		if output, err := git.CombinedOutput(); err != nil {
			return cleanup(fmt.Errorf("initialize Git repository: %w: %s", err, strings.TrimSpace(string(output))))
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project: %s\n\nNext steps:\n  cd %s\n  gecko dev\n", name, name)
	return nil
}

func isAvailableTemplate(name string) bool {
	for _, templateName := range availableTemplates {
		if name == templateName {
			return true
		}
	}
	return false
}

func extractTemplate(target, projectName, templateName string) error {
	return extractTemplateWithConfig(target, projectName, templateName, defaultSiteConfigValues(projectName))
}

func extractTemplateWithConfig(target, projectName, templateName string, values map[string]string) error {
	replacements := defaultSiteConfigValues(projectName)
	for token, value := range values {
		replacements[token] = value
	}
	source := filepath.ToSlash(filepath.Join("templates", templateName))
	return fs.WalkDir(starterTemplates, source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk template %q: %w", path, walkErr)
		}
		rel, err := filepath.Rel(filepath.FromSlash(source), filepath.FromSlash(path))
		if err != nil {
			return fmt.Errorf("resolve template path %q: %w", path, err)
		}
		if rel == "." {
			return nil
		}
		destination := filepath.Join(target, rel)
		relative, err := filepath.Rel(target, destination)
		if err != nil {
			return fmt.Errorf("resolve destination %q: %w", destination, err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
			return fmt.Errorf("unsafe template destination %q", destination)
		}
		if entry.IsDir() {
			if err := os.MkdirAll(destination, 0o755); err != nil {
				return fmt.Errorf("create directory %q: %w", rel, err)
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect template file %q: %w", rel, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported template entry type: %q", rel)
		}

		contents, err := starterTemplates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read template file %q: %w", rel, err)
		}
		if isTextTemplate(rel, contents) {
			fileReplacements := replacements
			if ext := strings.ToLower(filepath.Ext(rel)); ext == ".yaml" || ext == ".yml" {
				fileReplacements = make(map[string]string, len(replacements))
				for token, value := range replacements {
					quoted := strconv.Quote(value)
					fileReplacements[token] = quoted[1 : len(quoted)-1]
				}
			}
			contents = replaceTemplateTokens(contents, fileReplacements)
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return fmt.Errorf("create parent directory for %q: %w", rel, err)
		}
		if err := os.WriteFile(destination, contents, 0o644); err != nil {
			return fmt.Errorf("write template file %q: %w", rel, err)
		}
		return nil
	})
}

func isTextTemplate(path string, contents []byte) bool {
	if bytes.IndexByte(contents, 0) >= 0 {
		return false
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".html", ".htm", ".css", ".js", ".ts", ".json", ".yaml", ".yml", ".toml", ".xml", ".md", ".txt", ".go", ".svg":
		return true
	default:
		return false
	}
}

func replaceTemplateTokens(contents []byte, replacements map[string]string) []byte {
	tokens := make([]string, 0, len(replacements))
	for token := range replacements {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)
	arguments := make([]string, 0, len(tokens)*2)
	for _, token := range tokens {
		arguments = append(arguments, token, replacements[token])
	}
	return []byte(strings.NewReplacer(arguments...).Replace(string(contents)))
}

type siteConfigPrompt struct {
	Label, Token, Default string
}

func defaultSiteConfigValues(projectName string) map[string]string {
	title := siteTitle(projectName)
	return map[string]string{
		"__GECKO_SITE_TITLE__":       title,
		"__GECKO_SITE_DESCRIPTION__": "Personal portfolio, projects, experience, and writing.",
		"__GECKO_LANGUAGE__":         "en",
		"__GECKO_AUTHOR__":           "",
		"__GECKO_SITE_URL__":         "https://example.com",
		"__GECKO_BASE_URL__":         "",
		"__GECKO_POSTS_DIR__":        "content/posts",
		"__GECKO_TIMEZONE__":         "Asia/Kolkata",
	}
}

func promptSiteConfig(cmd *cobra.Command, projectName string) (map[string]string, error) {
	values := defaultSiteConfigValues(projectName)
	prompts := []siteConfigPrompt{
		{Label: "Site title", Token: "__GECKO_SITE_TITLE__", Default: values["__GECKO_SITE_TITLE__"]},
		{Label: "Site description", Token: "__GECKO_SITE_DESCRIPTION__", Default: values["__GECKO_SITE_DESCRIPTION__"]},
		{Label: "Language", Token: "__GECKO_LANGUAGE__", Default: values["__GECKO_LANGUAGE__"]},
		{Label: "Author", Token: "__GECKO_AUTHOR__", Default: values["__GECKO_AUTHOR__"]},
		{Label: "Site URL", Token: "__GECKO_SITE_URL__", Default: values["__GECKO_SITE_URL__"]},
		{Label: "Base URL prefix", Token: "__GECKO_BASE_URL__", Default: values["__GECKO_BASE_URL__"]},
		{Label: "Posts directory", Token: "__GECKO_POSTS_DIR__", Default: values["__GECKO_POSTS_DIR__"]},
		{Label: "Timezone", Token: "__GECKO_TIMEZONE__", Default: values["__GECKO_TIMEZONE__"]},
	}
	reader := bufio.NewReader(cmd.InOrStdin())
	writer := cmd.OutOrStdout()
	inputOpen := true
	for _, prompt := range prompts {
		if !inputOpen {
			continue
		}
		fmt.Fprintf(writer, "%s [%s]: ", prompt.Label, prompt.Default)
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("read %s: %w", prompt.Label, err)
		}
		if err == io.EOF {
			inputOpen = false
		}
		answer := strings.TrimSpace(line)
		if answer != "" {
			values[prompt.Token] = answer
		}
	}
	if err := validateSiteConfig(values); err != nil {
		return nil, err
	}
	return values, nil
}

func validateSiteConfig(values map[string]string) error {
	siteURL := values["__GECKO_SITE_URL__"]
	if siteURL != "" {
		parsed, err := url.ParseRequestURI(siteURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return fmt.Errorf("site URL must be an absolute HTTP or HTTPS URL")
		}
	}
	timezone := values["__GECKO_TIMEZONE__"]
	if _, err := time.LoadLocation(timezone); err != nil {
		return fmt.Errorf("invalid timezone %q: %w", timezone, err)
	}
	postsDir := values["__GECKO_POSTS_DIR__"]
	normalized := strings.ReplaceAll(postsDir, "\\", "/")
	path := filepath.FromSlash(normalized)
	if postsDir == "" || filepath.IsAbs(path) || filepath.VolumeName(path) != "" {
		return fmt.Errorf("posts directory must be a relative path inside the project")
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("posts directory must stay inside the project")
	}
	for _, r := range values["__GECKO_LANGUAGE__"] {
		if unicode.IsControl(r) {
			return fmt.Errorf("language must not contain control characters")
		}
	}
	return nil
}

func siteTitle(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' })
	for i, word := range words {
		runes := []rune(word)
		if len(runes) == 0 {
			continue
		}
		first := []rune(strings.ToUpper(string(runes[0])))
		if len(first) > 0 {
			runes = append(first, runes[1:]...)
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}
