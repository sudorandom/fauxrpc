package frontend

//go:generate go tool templ generate

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	mux.Handle("/sse/logs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := p.GetLogger()
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Send history
		history := logger.GetHistory()
		for _, entry := range history {
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			if err := enc.Encode(entry); err != nil {
				// TODO: handle error
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", buf.String())
		}
		flusher.Flush()

		ch, unsubscribe := logger.Subscribe()
		defer unsubscribe()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return // Client disconnected
			case entry := <-ch:
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				if err := enc.Encode(entry); err != nil {
					// TODO: handle error
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", buf.String())
				flusher.Flush()
			}
		}
	}))

	mux.Handle("/partials/request-log", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(partials.RequestLog()).ServeHTTP(w, r)
	}))
	mux.Handle("/partials/schema", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(partials.Schema()).ServeHTTP(w, r)
	}))
	mux.Handle("/partials/stubs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(partials.Stubs()).ServeHTTP(w, r)
	}))
	mux.Handle("/partials/summary", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(partials.Summary(p.GetStats())).ServeHTTP(w, r)
	}))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(templates.Index()).ServeHTTP(w, r)
	}))
	return mux
}
