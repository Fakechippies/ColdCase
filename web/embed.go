// Package web holds the embedded frontend assets for the ColdCase web UI:
// server-rendered HTML templates and static CSS/JS. Everything is compiled
// into the binary via go:embed so no Node build step or external files are
// required at runtime.
package web

import "embed"

// FS is the embedded asset filesystem.
//
//go:embed templates/*.html static/*
var FS embed.FS
