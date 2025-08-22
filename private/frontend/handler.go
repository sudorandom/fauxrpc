package frontend

//go:generate go tool templ generate
//go:generate bash -c "./generate_assets.sh"

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/sudorandom/fauxrpc/private/frontend/templates"
	"github.com/sudorandom/fauxrpc/private/frontend/templates/browser"
	"github.com/sudorandom/fauxrpc/private/frontend/templates/partials"
	"github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/metrics"
	"github.com/sudorandom/fauxrpc/private/registry"
	"google.golang.org/protobuf/reflect/protoreflect"
)

//go:embed assets
var embeddedAssets embed.FS

// Provider defines the interface for providing server statistics and the logger.
type Provider interface {
	GetStats() *metrics.Stats
	GetLogger() *log.Logger
	registry.ServiceRegistry
}

func DashboardHandler(p Provider) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/fauxrpc/sse/logs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := p.GetLogger()
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch, unsubscribe := logger.Subscribe()
		defer unsubscribe()

		ctx := r.Context()
		if _, err := fmt.Fprintf(w, "event: Connected\ndata: ok\n\n"); err != nil {
			slog.Error("failed to write new entry to SSE stream", "err", err)
			return
		}

		filter := strings.ToLower(r.URL.Query().Get("filter"))

		for {
			select {
			case <-ctx.Done():
				return // Client disconnected
			case entry := <-ch:
				// Apply filter if present
				if filter != "" {
					match := false
					switch {
					case strings.Contains(strings.ToLower(entry.Service), filter):
						match = true
					case strings.Contains(strings.ToLower(entry.Method), filter):
						match = true
					case strings.Contains(fmt.Sprintf("%d", entry.Status), filter):
						match = true
					case strings.Contains(strings.ToLower(string(entry.RequestBody)), filter):
						match = true
					case strings.Contains(strings.ToLower(string(entry.ResponseBody)), filter):
						match = true
					case strings.Contains(strings.ToLower(string(entry.RequestHeaders)), filter):
						match = true
					case strings.Contains(strings.ToLower(string(entry.ResponseHeaders)), filter):
						match = true
					}
					if !match {
						continue // Skip this entry if it doesn't match the filter
					}
				}

				var buf bytes.Buffer
				if err := partials.LogEntry(entry).Render(ctx, &buf); err != nil {
					slog.Error("failed to render log entry partial", "err", err)
					continue
				}

				html := buf.String()
				formattedHTML := strings.ReplaceAll(html, "\n", "\ndata: ")

				if _, err := fmt.Fprintf(w, "event: Request\ndata: %s\n\n", formattedHTML); err != nil {
					slog.Error("failed to write new entry to SSE stream", "err", err)
					return
				}
				flusher.Flush()
			}
		}
	}))

	mux.Handle("/fauxrpc/sse/summary", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return // Client disconnected
			case <-ticker.C:
				stats := p.GetStats()
				var buf bytes.Buffer
				if err := partials.Summary(stats).Render(ctx, &buf); err != nil {
					slog.Error("failed to render summary partial", "err", err)
					continue
				}
				if _, err := fmt.Fprintf(w, "event: Summary\ndata: %s\n\n", buf.String()); err != nil {
					slog.Error("failed to write summary to SSE stream", "err", err)
					return
				}
				flusher.Flush()
			}
		}
	}))

	mux.Handle("/fauxrpc/logs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "true" {
			// HTMX request, return only the partial
			templ.Handler(partials.RequestLog(nil)).ServeHTTP(w, r)
		} else {
			// Direct navigation, return full page with partial embedded
			templ.Handler(templates.Index(partials.RequestLog(nil))).ServeHTTP(w, r)
		}
	}))

	mux.Handle("/fauxrpc/browser/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/fauxrpc/browser")
		if after, ok := strings.CutPrefix(path, "/"); ok {
			path = after
		}
		files := p.Files()

		var foundFd protoreflect.FileDescriptor
		files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
			if fd.Path() == path {
				foundFd = fd
				return false
			}
			return true
		})

		if foundFd != nil {
			fd, err := desc.WrapFile(foundFd)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			printer := protoprint.Printer{}
			content, err := printer.PrintProtoToString(fd)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			lastSlash := strings.LastIndex(path, "/")
			var dirPath, fileName string
			if lastSlash != -1 {
				dirPath = path[:lastSlash+1]
				fileName = path[lastSlash+1:]
			} else {
				dirPath = ""
				fileName = path
			}

			if r.Header.Get("HX-Request") == "true" {
				templ.Handler(browser.FileContent(dirPath, fileName, content)).ServeHTTP(w, r)
			} else {
				templ.Handler(templates.Index(browser.FileContent(dirPath, fileName, content))).ServeHTTP(w, r)
			}
			return
		}

		// It's a directory
		dirEntries := make(map[string]bool)
		files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
			if after, ok := strings.CutPrefix(fd.Path(), path); ok {
				rest := after
				if rest == "" {
					return true
				}
				parts := strings.Split(rest, "/")
				if len(parts) == 1 {
					dirEntries[parts[0]] = true
				} else {
					dirEntries[parts[0]+"/"] = true
				}
			}
			return true
		})

		var entries []string
		for entry := range dirEntries {
			entries = append(entries, entry)
		}
		sort.Strings(entries)

		if r.Header.Get("HX-Request") == "true" {
			templ.Handler(browser.Browser(path, entries)).ServeHTTP(w, r)
		} else {
			templ.Handler(templates.Index(browser.Browser(path, entries))).ServeHTTP(w, r)
		}
	}))

	mux.Handle("/fauxrpc/summary", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "true" {
			// HTMX request, return only the partial
			templ.Handler(partials.SummaryPage(p.GetStats())).ServeHTTP(w, r)
		} else {
			// Direct navigation, return full page with partial embedded
			templ.Handler(templates.Index(partials.SummaryPage(p.GetStats()))).ServeHTTP(w, r)
		}
	}))

	mux.Handle("/fauxrpc/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(templates.Index(partials.SummaryPage(p.GetStats()))).ServeHTTP(w, r)
	}))

	// Serve static assets from the embedded file system
	assetFS, err := fs.Sub(embeddedAssets, "assets")
	if err != nil {
		slog.Error("failed to create sub-filesystem for assets", "err", err)
		// Handle error appropriately, perhaps return a handler that always errors
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal server error: asset loading failed", http.StatusInternalServerError)
		})
	}
	mux.Handle("/fauxrpc/assets/", http.StripPrefix("/fauxrpc/assets/", http.FileServer(http.FS(assetFS))))

	return mux
}
