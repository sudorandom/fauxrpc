package frontend

//go:generate go tool templ generate

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/sudorandom/fauxrpc/private/frontend/templates"
	"github.com/sudorandom/fauxrpc/private/frontend/templates/partials"
	"github.com/sudorandom/fauxrpc/private/metrics" // Added import
)

// StatsProvider defines the interface for providing server statistics.
type StatsProvider interface {
	GetStats() *metrics.Stats // Updated to metrics.Stats
}

func DashboardHandler(s StatsProvider) http.Handler {
	mux := http.NewServeMux()
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
		templ.Handler(partials.Summary(s.GetStats())).ServeHTTP(w, r)
	}))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(templates.Index()).ServeHTTP(w, r)
	}))
	return mux
}
