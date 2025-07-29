# FauxRPC Frontend

This directory contains the React dashboard UI for FauxRPC, along with the Go embed code.

## Development

- Edit the UI in `src/dashboard.tsx`.
- Entry point is `src/main.tsx`.
- Run `npm install` in this directory to install dependencies.
- Run `npm run build` to build production assets to `dashboard_dist/`.

## Embedding in Go

The Go server uses `embed.FS` to serve the built assets from `dashboard_dist/`.

## Go Generate

Add the following to your Go file (already present in dashboard_embed.go):

```
//go:generate npm --prefix ./private/frontend install
//go:generate npm --prefix ./private/frontend run build
```

This will ensure assets are rebuilt and embedded when you run `go generate ./...`.
