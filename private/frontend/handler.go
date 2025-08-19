package frontend

//go:generate go tool templ generate

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/sudorandom/fauxrpc/private/frontend/templates"
	"github.com/sudorandom/fauxrpc/private/frontend/templates/partials"
	"github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/metrics"
)

// Provider defines the interface for providing server statistics and the logger.
type Provider interface {
	GetStats() *metrics.Stats
	GetLogger() *log.Logger
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

		fmt.Println("filter", filter)

		for {
			select {
			case <-ctx.Done():
				return // Client disconnected
			case entry := <-ch:
				// Apply filter if present
				if filter != "" {
					match := false
					if strings.Contains(strings.ToLower(entry.Service), filter) {
						match = true
					}
					if !match && strings.Contains(strings.ToLower(entry.Method), filter) {
						match = true
					}
					if !match && strings.Contains(fmt.Sprintf("%d", entry.Status), filter) {
						match = true
					}
					// You can add more fields to filter here, e.g., headers, body content
					// For example:
					// if !match && strings.Contains(strings.ToLower(string(entry.RequestBody)), filter) {
					// 	match = true
					// }
					// if !match && strings.Contains(strings.ToLower(string(entry.ResponseBody)), filter) {
					// 	match = true
					// }

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
	return mux
}
