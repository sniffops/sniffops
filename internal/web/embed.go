package web

import (
	"embed"
	"io/fs"
)

// DistFS embeds the web/dist/ directory (frontend build output)
// 
// This uses Go 1.16+ embed directive to bundle the React/Preact frontend
// into the binary, enabling single-file distribution.
//
// Build process:
// 1. cd web && npm run build  -> creates web/dist/
// 2. go build cmd/sniffops/main.go -> embeds web/dist/ into binary
//
// The embedded filesystem is served at the root path "/" by server.go

//go:embed dist
var distFS embed.FS

// DistFS is the public filesystem interface for serving embedded files
var DistFS fs.FS

func init() {
	// Strip "dist/" prefix when serving
	// This allows files in web/dist/index.html to be served at /index.html
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		// If dist/ doesn't exist (e.g., during development without frontend build),
		// use an empty filesystem to avoid build errors
		DistFS = emptyFS{}
	} else {
		DistFS = sub
	}
}

// emptyFS is a fallback filesystem when dist/ directory doesn't exist
type emptyFS struct{}

func (emptyFS) Open(name string) (fs.File, error) {
	return nil, fs.ErrNotExist
}
