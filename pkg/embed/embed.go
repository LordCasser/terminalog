// Package embed provides static file embedding for the frontend.
package embed

import (
	"embed"
	"io/fs"
)

// StaticFS embeds the static frontend files.
// The static directory contains the Next.js build output.
//
//go:embed all:static
var StaticFS embed.FS

// GetStaticFS returns the embedded static filesystem.
func GetStaticFS() embed.FS {
	return StaticFS
}

// GetStaticSubFS returns the static subdirectory as a filesystem.
func GetStaticSubFS() (fs.FS, error) {
	return fs.Sub(StaticFS, "static")
}
