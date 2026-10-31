/*
Copyright © 2026 dos <debasmitbos22@gmail.com>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

const liveReloadPath = "/__dev/events"

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Build and serve the site with live updates",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		directory, _ := cmd.Flags().GetString("dir")
		output, _ := cmd.Flags().GetString("output")
		port, _ := cmd.Flags().GetInt("port")
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		return runDev(ctx, directory, output, port, cmd.OutOrStdout())
	},
}

func init() {
	devCmd.Flags().String("dir", ".", "Project directory")
	devCmd.Flags().Int("port", 3000, "Development server port")
	devCmd.Flags().String("output", "public", "Output directory")
	rootCmd.AddCommand(devCmd)
}

func runDev(ctx context.Context, directory, output string, port int, logOutput interface{ Write([]byte) (int, error) }) error {
	root, err := filepath.Abs(directory)
	if err != nil {
		return fmt.Errorf("resolve project directory: %w", err)
	}
	outputPath := output
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(root, outputPath)
	}
	outputPath, err = filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if err := validateDevOutput(root, outputPath); err != nil {
		return err
	}

	staging := filepath.Join(root, ".gecko-dev-staging")
	if err := rebuildAndPublish(root, staging, outputPath); err != nil {
		return fmt.Errorf("initial build: %w", err)
	}
	fmt.Fprintf(logOutput, "\n  Gecko development server\n\n  ✓ Initial build completed\n\n  Local:    http://localhost:%d\n  Watching: content/, layouts/, templates/, assets/\n\n", port)

	hub := newReloadHub()
	server := &http.Server{Addr: fmt.Sprintf("localhost:%d", port), Handler: devHandler(outputPath, hub)}
	serverErrors := make(chan error, 1)
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	go func() { serverErrors <- server.Serve(listener) }()
	defer server.Shutdown(context.Background())

	watcher, err := newProjectWatcher(root)
	if err != nil {
		return err
	}
	defer watcher.Close()
	return watchAndRebuild(ctx, watcher, root, staging, outputPath, hub, logOutput, serverErrors)
}

func validateDevOutput(root, output string) error {
	if output == root {
		return fmt.Errorf("output directory must not be the project directory")
	}
	for _, source := range []string{"content", "layouts", "templates", "assets", "static"} {
		if isWithin(output, filepath.Join(root, source)) {
			return fmt.Errorf("output directory must not be inside %s/", source)
		}
	}
	return nil
}

func rebuildAndPublish(root, staging, output string) error {
	if err := buildSiteTo(root, staging); err != nil {
		return err
	}
	backup := output + ".gecko-previous"
	_ = os.RemoveAll(backup)
	if _, err := os.Stat(output); err == nil {
		if err := os.Rename(output, backup); err != nil {
			return fmt.Errorf("stage existing output: %w", err)
		}
	}
	if err := os.Rename(staging, output); err != nil {
		if _, restoreErr := os.Stat(backup); restoreErr == nil {
			_ = os.Rename(backup, output)
		}
		return fmt.Errorf("publish build: %w", err)
	}
	return os.RemoveAll(backup)
}

func newProjectWatcher(root string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	for _, source := range []string{"content", "layouts", "templates", "assets", "static"} {
		if err := addWatchTree(watcher, filepath.Join(root, source)); err != nil {
			watcher.Close()
			return nil, err
		}
	}
	if err := watcher.Add(root); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("watch project root for site.yaml: %w", err)
	}
	return watcher, nil
}

func addWatchTree(watcher *fsnotify.Watcher, directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return filepath.WalkDir(directory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
}

func watchAndRebuild(ctx context.Context, watcher *fsnotify.Watcher, root, staging, output string, hub *reloadHub, logOutput interface{ Write([]byte) (int, error) }, serverErrors <-chan error) error {
	var timer *time.Timer
	var timerC <-chan time.Time
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-serverErrors:
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return fmt.Errorf("development server: %w", err)
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(logOutput, "Watch error: %v\n", err)
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if filepath.Dir(event.Name) == root && filepath.Base(event.Name) != "site.yaml" {
				continue
			}
			if !shouldRebuild(event) {
				continue
			}
			if event.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					_ = addWatchTree(watcher, event.Name)
				}
			}
			if timer == nil {
				timer = time.NewTimer(150 * time.Millisecond)
				timerC = timer.C
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(150 * time.Millisecond)
			}
		case <-timerC:
			timer = nil
			timerC = nil
			fmt.Fprintln(logOutput, "Rebuilding...")
			if err := rebuildAndPublish(root, staging, output); err != nil {
				fmt.Fprintf(logOutput, "Build failed: %v\n", err)
				continue
			}
			fmt.Fprintln(logOutput, "✓ Build completed")
			hub.broadcast()
		}
	}
}

func shouldRebuild(event fsnotify.Event) bool {
	name := filepath.Base(event.Name)
	return !strings.HasPrefix(name, ".") && !strings.HasSuffix(name, "~") && event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0
}

func isWithin(pathname, directory string) bool {
	rel, err := filepath.Rel(directory, pathname)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

type reloadHub struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

func newReloadHub() *reloadHub { return &reloadHub{clients: make(map[chan struct{}]struct{})} }
func (hub *reloadHub) broadcast() {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	for client := range hub.clients {
		select {
		case client <- struct{}{}:
		default:
		}
	}
}
func (hub *reloadHub) serveHTTP(response http.ResponseWriter, request *http.Request) {
	flusher, ok := response.(http.Flusher)
	if !ok {
		http.Error(response, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Connection", "keep-alive")
	client := make(chan struct{}, 1)
	hub.mu.Lock()
	hub.clients[client] = struct{}{}
	hub.mu.Unlock()
	defer func() { hub.mu.Lock(); delete(hub.clients, client); hub.mu.Unlock() }()
	fmt.Fprint(response, "event: ready\ndata: connected\n\n")
	flusher.Flush()
	for {
		select {
		case <-request.Context().Done():
			return
		case <-client:
			fmt.Fprint(response, "event: reload\ndata: updated\n\n")
			flusher.Flush()
		}
	}
}

func devHandler(output string, hub *reloadHub) http.Handler {
	files := http.FileServer(http.Dir(output))
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == liveReloadPath {
			hub.serveHTTP(response, request)
			return
		}
		name := path.Clean("/" + request.URL.Path)
		local := filepath.Join(output, filepath.FromSlash(strings.TrimPrefix(name, "/")))
		if info, err := os.Stat(local); err == nil && info.IsDir() {
			local = filepath.Join(local, "index.html")
		}
		if strings.HasSuffix(local, ".html") {
			if contents, err := os.ReadFile(local); err == nil {
				response.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = response.Write(injectLiveReload(contents))
				return
			}
		}
		files.ServeHTTP(response, request)
	})
}

func injectLiveReload(contents []byte) []byte {
	script := []byte(`<script>new EventSource("/__dev/events").addEventListener("reload",function(){window.location.reload()})</script>`)
	if position := bytesLastIndex(contents, []byte("</body>")); position >= 0 {
		return append(append(contents[:position], script...), contents[position:]...)
	}
	return append(contents, script...)
}

func bytesLastIndex(data, needle []byte) int { return strings.LastIndex(string(data), string(needle)) }
