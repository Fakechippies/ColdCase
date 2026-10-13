// Package server implements the ColdCase local web UI. It serves
// server-rendered HTML from embedded templates and exposes a small JSON/SSE
// API for streaming tool execution. It reuses pkg/catalog for the tool
// catalog, pkg/runner for execution and pkg/session for audit state, so the
// web surface stays consistent with the CLI.
package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"coldcase/pkg/catalog"
	"coldcase/pkg/session"
	"coldcase/web"
)

// csrfCookie is the name of the double-submit CSRF cookie.
const csrfCookie = "coldcase_csrf"

// csrfContextKey is the context key holding the active CSRF token.
type csrfContextKey struct{}

// Server is the ColdCase web UI server.
type Server struct {
	addr    string
	pages   map[string]*template.Template
	mux     *http.ServeMux
	manager *session.Manager

	toolCount     int
	categoryCount int
}

// New constructs a Server. addr is the full listen address, e.g.
// "127.0.0.1:8765".
func New(addr string) (*Server, error) {
	pages, err := parseTemplates()
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}

	s := &Server{
		addr:    addr,
		pages:   pages,
		mux:     http.NewServeMux(),
		manager: session.NewManager(),
	}
	for _, c := range catalog.Categories() {
		s.categoryCount++
		s.toolCount += len(c.Tools)
	}
	s.routes()
	return s, nil
}

// Handler returns the root HTTP handler (middleware included).
func (s *Server) Handler() http.Handler {
	return s.withCommon(s.mux)
}

// Addr returns the configured listen address.
func (s *Server) Addr() string { return s.addr }

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return srv.ListenAndServe()
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /{$}", s.handleDashboard)
	s.mux.HandleFunc("GET /tools", s.handleTools)
	s.mux.HandleFunc("GET /tools/{name}", s.handleTool)
	s.mux.HandleFunc("GET /status", s.handleStatus)
	s.mux.HandleFunc("GET /sessions", s.handleSessions)
	s.mux.HandleFunc("POST /sessions", s.handleCreateSession)
	s.mux.HandleFunc("GET /sessions/{id}", s.handleSession)
	s.mux.HandleFunc("POST /sessions/{id}/action", s.handleSessionAction)
	s.mux.HandleFunc("GET /sessions/{id}/export", s.handleExport)
	s.mux.HandleFunc("POST /keys/generate", s.handleKeyGenerate)
	s.mux.HandleFunc("POST /api/run", s.handleRun)
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(mustSub(web.FS, "static")))))
}

// ─── middleware ─────────────────────────────────────────────────────────────

func (s *Server) withCommon(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers. CSP is strict: no inline scripts/styles.
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("Content-Security-Policy",
			"default-src 'self'; img-src 'self' data:; style-src 'self'; "+
				"script-src 'self'; connect-src 'self'; base-uri 'none'; "+
				"form-action 'self'; frame-ancestors 'none'")

		// Double-submit CSRF cookie, issued on first request. The resolved
		// token is placed in the request context so the very first page can
		// render a usable form even before the client echoes the cookie.
		token := ""
		if c, err := r.Cookie(csrfCookie); err == nil {
			token = c.Value
		} else {
			token, err = randomToken()
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     csrfCookie,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
		}
		r = r.WithContext(context.WithValue(r.Context(), csrfContextKey{}, token))

		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			h.Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

// checkCSRF validates the double-submit token for unsafe requests. The token
// may be supplied either as the X-CSRF-Token header (fetch) or a "csrf" form
// field (HTML form).
func (s *Server) checkCSRF(r *http.Request) error {
	c, err := r.Cookie(csrfCookie)
	if err != nil {
		return errors.New("missing CSRF cookie")
	}
	got := r.Header.Get("X-CSRF-Token")
	if got == "" {
		if err := r.ParseForm(); err == nil {
			got = r.FormValue("csrf")
		}
	}
	if subtle.ConstantTimeCompare([]byte(c.Value), []byte(got)) != 1 {
		return errors.New("invalid CSRF token")
	}
	return nil
}

// ─── template plumbing ──────────────────────────────────────────────────────

func parseTemplates() (map[string]*template.Template, error) {
	funcs := template.FuncMap{
		"fmtTime": func(t time.Time) string { return t.Format("2006-01-02 15:04:05") },
		"shortHash": func(s string) string {
			if len(s) > 16 {
				return s[:16] + "…"
			}
			return s
		},
		"joinArgs": func(a []string) string { return strings.Join(a, " ") },
		"slug": func(s string) string {
			var b strings.Builder
			for _, r := range strings.ToLower(s) {
				switch {
				case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
					b.WriteRune(r)
				case r == ' ' || r == '-' || r == '_' || r == '&':
					b.WriteRune('-')
				}
			}
			return strings.Trim(b.String(), "-")
		},
	}

	base, err := template.New("layout.html").Funcs(funcs).
		ParseFS(web.FS, "templates/layout.html", "templates/partials.html")
	if err != nil {
		return nil, err
	}

	entries, err := fs.ReadDir(web.FS, "templates")
	if err != nil {
		return nil, err
	}

	pages := make(map[string]*template.Template)
	for _, e := range entries {
		name := e.Name()
		if name == "layout.html" || name == "partials.html" {
			continue
		}
		clone, err := base.Clone()
		if err != nil {
			return nil, err
		}
		if _, err := clone.ParseFS(web.FS, "templates/"+name); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		pages[name] = clone
	}
	return pages, nil
}

// render executes a page template wrapped in the shared layout.
func (s *Server) render(w http.ResponseWriter, page string, data *viewData) {
	t, ok := s.pages[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout.html", data); err != nil {
		// Headers may already be written; log to the server's stderr.
		fmt.Printf("[web] template error (%s): %v\n", page, err)
	}
}

// renderError writes a full HTML error page with the given status.
func (s *Server) renderError(w http.ResponseWriter, r *http.Request, code int, title, detail string, args ...any) {
	if len(args) > 0 {
		detail = fmt.Sprintf(detail, args...)
	}
	w.WriteHeader(code)
	s.render(w, "error.html", &viewData{
		Title:         "Error",
		Nav:           "",
		CSRF:          CSRFToken(r),
		ErrorCode:     fmt.Sprintf("%d", code),
		ErrorTitle:    title,
		ErrorDetail:   detail,
		ToolCount:     s.toolCount,
		CategoryCount: s.categoryCount,
	})
}

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
