# Repository Guidelines

## Project Structure

This is the Go-based `gecko` static site generator. `main.go` starts the CLI, and `cmd/` contains Cobra commands and their tests. The embedded starter project lives in `cmd/templates/portfolio/`: Markdown content, Go `html/template` layouts, CSS, and static files such as `favicon.ico`. Keep new commands in focused files (for example, `cmd/clean.go`) and tests beside their implementation (`*_test.go`).

## Build, Test, and Development

- `go run . --help` lists available commands and flags.
- `go run . init my-site` creates a project from the embedded template.
- `go run . build` generates production files in `dist/`.
- `go run . dev` builds, serves, and watches a project; use `--dir`, `--port`, and `--output` to configure it.
- `go run . preview` builds and serves the generated site.
- `go test ./...` runs unit and command tests; `go vet ./...` checks for common Go issues.
- `gofmt -w cmd/*.go` formats command code.

Shell completions are generated with `gecko completion bash`, `gecko completion zsh`, or `gecko completion fish`; see root help for setup examples.

## Style and Naming

Use idiomatic Go, format with `gofmt`, and keep package names lowercase. Exported identifiers use `PascalCase`; unexported identifiers use `camelCase`. Name commands and files descriptively (`new.go`, `preview.go`) and tests `*_test.go` with behavior-focused `Test...` names. Return actionable errors from Cobra `RunE` handlers and write command output through `cmd.OutOrStdout()`.

## Template and Content Changes

Keep starter paths stable because `gecko init` embeds and copies the entire portfolio tree. HTML layouts use Go `html/template` blocks; content uses Markdown with YAML front matter. Put the default site icon at `cmd/templates/portfolio/static/favicon.ico` so builds copy it to the published site root. Test changes to template output through the shared build pipeline.

## Commits and Pull Requests

No Git history is available here to establish a commit convention. Use concise imperative subjects, such as `Add fish completion guidance`. Pull requests should describe user-visible changes, list validation commands, and include terminal output or screenshots when CLI help or rendered pages change.
