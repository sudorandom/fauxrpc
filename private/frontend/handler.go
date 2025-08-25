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
	templates_stubs "github.com/sudorandom/fauxrpc/private/frontend/templates/stubs"
	"github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/metrics"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"google.golang.org/protobuf/reflect/protoreflect"

	stubsv1 "github.com/sudorandom/fauxrpc/private/gen/stubs/v1"
)

//go:embed assets
var embeddedAssets embed.FS

// Provider defines the interface for providing server statistics and the logger.
type Provider interface {
	GetStats() *metrics.Stats
	GetLogger() *log.Logger
	registry.ServiceRegistry
	GetStubDB() stubs.StubDatabase
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

		// Check if it's a file
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

			pkgName := string(foundFd.Package())
			lastSlash := strings.LastIndex(foundFd.Path(), "/")
			var fileName string
			if lastSlash != -1 {
				fileName = foundFd.Path()[lastSlash+1:]
			} else {
				fileName = foundFd.Path()
			}

			if r.Header.Get("HX-Request") == "true" {
				templ.Handler(browser.FileContent(pkgName, fileName, content)).ServeHTTP(w, r)
			} else {
				templ.Handler(templates.Index(browser.FileContent(pkgName, fileName, content))).ServeHTTP(w, r)
			}
			return
		}

		// Check if it's a package
		var fds []protoreflect.FileDescriptor
		files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
			if string(fd.Package()) == path {
				fds = append(fds, fd)
			}
			return true
		})

		if len(fds) > 0 {
			var entries []string
			for _, fd := range fds {
				entries = append(entries, fd.Path())
			}
			sort.Strings(entries)
			if r.Header.Get("HX-Request") == "true" {
				templ.Handler(browser.Browser(path, entries)).ServeHTTP(w, r)
			} else {
				templ.Handler(templates.Index(browser.Browser(path, entries))).ServeHTTP(w, r)
			}
			return
		}

		// It's the root, show packages
		packages := make(map[string]bool)
		files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
			pkg := string(fd.Package())
			if pkg != "" {
				packages[pkg] = true
			}
			return true
		})

		var entries []string
		for pkg := range packages {
			entries = append(entries, pkg)
		}
		sort.Strings(entries)

		if r.Header.Get("HX-Request") == "true" {
			templ.Handler(browser.Browser("", entries)).ServeHTTP(w, r)
		} else {
			templ.Handler(templates.Index(browser.Browser("", entries))).ServeHTTP(w, r)
		}
	}))

	mux.Handle("/fauxrpc/stubs/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/fauxrpc/stubs")
		if after, ok := strings.CutPrefix(path, "/"); ok {
			path = after
		}

		// Handle single stub view
		if path != "" {
			parts := strings.Split(path, "/")
			if len(parts) == 2 {
				target := parts[0]
				id := parts[1]

				stubEntry, ok := p.GetStubDB().GetStub(stubs.StubKey{Name: protoreflect.FullName(target), ID: id})
				if !ok {
					http.Error(w, "Stub not found", http.StatusNotFound)
					return
				}

				pbStubs, err := stubs.StubsToProto([]stubs.StubEntry{stubEntry})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if len(pbStubs) == 0 {
					http.Error(w, "Stub not found after conversion", http.StatusNotFound)
					return
				}

				if r.Header.Get("HX-Request") == "true" {
					templ.Handler(templates_stubs.Single(pbStubs[0])).ServeHTTP(w, r)
				} else {
					templ.Handler(templates.Index(templates_stubs.Single(pbStubs[0]))).ServeHTTP(w, r)
				}
				return
			}
		}

		// Handle list all stubs
		allStubEntries := p.GetStubDB().GetStubs()
		pbStubs, err := stubs.StubsToProto(allStubEntries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		groupedStubs := make(map[string][]*stubsv1.Stub)
		for _, stub := range pbStubs {
			groupedStubs[stub.GetRef().GetTarget()] = append(groupedStubs[stub.GetRef().GetTarget()], stub)
		}

		// Sort targets
		var targets []string
		for target := range groupedStubs {
			targets = append(targets, target)
		}
		sort.Strings(targets)

		// Sort stubs within each target and create an ordered map
		orderedGroupedStubs := make(map[string][]*stubsv1.Stub)
		for _, target := range targets {
			stubs := groupedStubs[target]
			sort.Slice(stubs, func(i, j int) bool {
				return stubs[i].GetRef().GetId() < stubs[j].GetRef().GetId()
			})
			orderedGroupedStubs[target] = stubs
		}

		if r.Header.Get("HX-Request") == "true" {
			templ.Handler(templates_stubs.List(orderedGroupedStubs)).ServeHTTP(w, r)
		} else {
			templ.Handler(templates.Index(templates_stubs.List(orderedGroupedStubs))).ServeHTTP(w, r)
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

	mux.Handle("/fauxrpc/about", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "true" {
			templ.Handler(templates.About()).ServeHTTP(w, r)
		} else {
			templ.Handler(templates.Index(templates.About())).ServeHTTP(w, r)
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
