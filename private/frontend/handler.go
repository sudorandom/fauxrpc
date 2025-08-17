package frontend

//go:generate go tool templ generate

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"

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

		slog.Info("SSE handler started, listening for new entries")
		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				slog.Info("SSE client disconnected")
				return // Client disconnected
			case entry := <-ch:
				var buf bytes.Buffer
				if err := partials.LogEntry(entry).Render(ctx, &buf); err != nil {
					slog.Error("failed to render log entry partial", "err", err)
					continue
				}
				if _, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", buf.String()); err != nil {
					slog.Error("failed to write new entry to SSE stream", "err", err)
					return
				}
				flusher.Flush()
			}
		}
	}))

	mux.Handle("/fauxrpc/partials/request-log", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(partials.RequestLog(nil)).ServeHTTP(w, r)
	}))

	mux.Handle("/fauxrpc/partials/schema", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(partials.Schema()).ServeHTTP(w, r)
	}))
	mux.Handle("/fauxrpc/partials/stubs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(partials.Stubs()).ServeHTTP(w, r)
	}))
	mux.Handle("/fauxrpc/partials/summary", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(partials.Summary(p.GetStats())).ServeHTTP(w, r)
	}))
	mux.Handle("/fauxrpc/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(templates.Index()).ServeHTTP(w, r)
	}))
	return mux
}
