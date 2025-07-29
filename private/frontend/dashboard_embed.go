//go:generate npm install
//go:generate npm run build

package frontend

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dashboard_dist/**
var dashboardFS embed.FS

func DashboardHandler() http.Handler {
	f, _ := fs.Sub(dashboardFS, "dashboard_dist")
	return http.StripPrefix("/dashboard", http.FileServer(http.FS(f)))
}
