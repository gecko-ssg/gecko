package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpAndVersionFlags(t *testing.T) {
	for _, flag := range []string{"help", "version"} {
		if rootCmd.Flags().Lookup(flag) == nil {
			t.Errorf("root flag --%s was not registered", flag)
		}
	}
	if rootCmd.Flags().ShorthandLookup("h") == nil {
		t.Error("root help shorthand -h was not registered")
	}
	if rootCmd.Flags().ShorthandLookup("v") == nil {
		t.Error("root version shorthand -v was not registered")
	}
	for _, name := range []string{"build", "clean", "completion", "dev", "init", "new", "preview", "version"} {
		if _, _, err := rootCmd.Find([]string{name}); err != nil {
			t.Errorf("command %q is not discoverable: %v", name, err)
		}
	}
}

func TestShellCompletionGeneration(t *testing.T) {
	generators := []struct {
		name string
		gen  func(*bytes.Buffer) error
	}{
		{"bash", func(out *bytes.Buffer) error { return rootCmd.GenBashCompletionV2(out, true) }},
		{"zsh", func(out *bytes.Buffer) error { return rootCmd.GenZshCompletion(out) }},
		{"fish", func(out *bytes.Buffer) error { return rootCmd.GenFishCompletion(out, true) }},
	}
	for _, generator := range generators {
		t.Run(generator.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := generator.gen(&output); err != nil {
				t.Fatal(err)
			}
			if output.Len() == 0 || !strings.Contains(output.String(), "gecko") {
				t.Fatalf("invalid %s completion output", generator.name)
			}
		})
	}
}
