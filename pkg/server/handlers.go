package server

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"coldcase/pkg/catalog"
	"coldcase/pkg/runner"
	"coldcase/pkg/session"
	"coldcase/pkg/tools"
)

var sessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// ─── pages ──────────────────────────────────────────────────────────────────

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	d := s.newData("Dashboard", "dashboard", r)
	d.Subtitle = "Local-only forensics console"

	sums := s.loadSummaries()
	d.SessionCount = len(sums)
	if len(sums) > 5 {
		sums = sums[:5]
	}
	d.RecentSessions = sums

	seen := map[string]bool{}
	for _, t := range catalog.All() {
		if seen[t.Binary] {
			continue
		}
		seen[t.Binary] = true
		if tools.CheckToolInstalled(t.Binary) {
			d.NativeCount++
		} else if d.ContainerReady {
			d.ContainerCount++
		}
	}
	s.render(w, "dashboard.html", d)
}

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("q"))
	q := strings.ToLower(raw)

	d := s.newData("Tool Catalog", "tools", r)
	d.Subtitle = fmt.Sprintf("%d tools across %d categories", s.toolCount, s.categoryCount)
	d.Query = raw
	d.Availability = map[string]availabilityView{}

	for _, c := range catalog.Categories() {
		var kept []catalog.Tool
		for _, t := range c.Tools {
			av := runner.CheckAvailability(t.Binary)
			d.Availability[t.Name] = availabilityView{Native: av.Native, Container: av.Container}
			if q == "" || strings.Contains(strings.ToLower(t.Name+" "+t.Description+" "+t.Binary), q) {
				kept = append(kept, t)
			}
		}
		if len(kept) > 0 {
			d.Categories = append(d.Categories, catalog.Category{Name: c.Name, Tools: kept})
		}
	}
	s.render(w, "tools.html", d)
}

func (s *Server) handleTool(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	t, ok := catalog.Find(name)
	if !ok {
		s.renderError(w, r, http.StatusNotFound, "Tool not found",
			"No tool named %q exists in the catalog.", name)
		return
	}
	av := runner.CheckAvailability(t.Binary)

	d := s.newData(t.Name, "tools", r)
	d.Subtitle = t.Category
	d.Tool = t
	d.Avail = availabilityView{Native: av.Native, Container: av.Container}
	d.SessionOptions = s.loadSessionOptions("")
	s.render(w, "tool.html", d)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	d := s.newData("Availability", "status", r)
	d.Subtitle = "Native and container availability for every catalog tool"

	for _, t := range catalog.All() {
		av := runner.CheckAvailability(t.Binary)
		row := statusRow{
			Name:           t.Name,
			Category:       t.Category,
			Binary:         t.Binary,
			Native:         av.Native,
			ContainerReady: av.Container,
			Available:      av.Available(),
		}
		if row.Available {
			d.AvailableCount++
		}
		d.Rows = append(d.Rows, row)
	}
	s.render(w, "status.html", d)
}

// ─── sessions ───────────────────────────────────────────────────────────────

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	d := s.newData("Sessions", "sessions", r)
	d.Subtitle = "Signed chain-of-custody for tool executions"
	d.Sessions = s.loadSummaries()
	s.render(w, "sessions.html", d)
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	if err := s.checkCSRF(r); err != nil {
		s.renderError(w, r, http.StatusForbidden, "Forbidden", "%s", err.Error())
		return
	}
	_ = r.ParseForm()

	id := strings.TrimSpace(r.FormValue("id"))
	if !sessionIDPattern.MatchString(id) {
		s.renderError(w, r, http.StatusBadRequest, "Invalid session ID",
			"Use 1–64 characters: letters, digits, dot, underscore or hyphen.")
		return
	}
	investigator := strings.TrimSpace(r.FormValue("investigator"))
	email := strings.TrimSpace(r.FormValue("email"))
	sign := r.FormValue("sign") != ""
	encrypt := r.FormValue("encrypt") != ""

	if _, err := s.manager.Create(id, investigator, email, sign, encrypt); err != nil {
		s.renderError(w, r, http.StatusBadRequest, "Could not create session", "%s", err.Error())
		return
	}
	http.Redirect(w, r, "/sessions/"+id, http.StatusSeeOther)
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess, err := s.manager.Load(id)
	if err != nil {
		s.renderError(w, r, http.StatusNotFound, "Session not found",
			"No session named %q exists.", id)
		return
	}

	d := s.newData("Session "+sess.ID, "sessions", r)
	d.Subtitle = fmt.Sprintf("State: %s · %d command(s)", sess.State, len(sess.Commands))
	d.Session = sess
	d.Commands = s.buildCommandViews(sess)
	d.VerifyState, d.VerifyMsg = s.verifySession(sess)
	s.render(w, "session.html", d)
}

func (s *Server) handleSessionAction(w http.ResponseWriter, r *http.Request) {
	if err := s.checkCSRF(r); err != nil {
		s.renderError(w, r, http.StatusForbidden, "Forbidden", "%s", err.Error())
		return
	}
	_ = r.ParseForm()

	id := r.PathValue("id")
	sess, err := s.manager.Load(id)
	if err != nil {
		s.renderError(w, r, http.StatusNotFound, "Session not found", "No session named %q exists.", id)
		return
	}

	switch action := r.FormValue("action"); action {
	case "lock":
		err = s.manager.Lock(sess)
	case "unlock":
		err = s.manager.Unlock(sess)
	case "seal":
		err = s.manager.Seal(sess)
	case "verify":
		// Verification is computed on render; nothing to persist.
	default:
		s.renderError(w, r, http.StatusBadRequest, "Unknown action", "Unsupported action %q.", action)
		return
	}
	if err != nil {
		s.renderError(w, r, http.StatusBadRequest, "Action failed", "%s", err.Error())
		return
	}
	http.Redirect(w, r, "/sessions/"+id, http.StatusSeeOther)
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess, err := s.manager.Load(id)
	if err != nil {
		s.renderError(w, r, http.StatusNotFound, "Session not found", "No session named %q exists.", id)
		return
	}

	var body, contentType, ext string
	switch format := r.URL.Query().Get("format"); format {
	case "json":
		body, err = session.ExportJSON(sess)
		contentType, ext = "application/json", "json"
	case "markdown", "md":
		body, err = session.ExportMarkdown(sess)
		contentType, ext = "text/markdown; charset=utf-8", "md"
	case "html":
		body, err = session.ExportHTML(sess)
		contentType, ext = "text/html; charset=utf-8", "html"
	default:
		s.renderError(w, r, http.StatusBadRequest, "Unknown format",
			"Supported formats: json, markdown, html.")
		return
	}
	if err != nil {
		s.renderError(w, r, http.StatusInternalServerError, "Export failed", "%s", err.Error())
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", sess.ID+"."+ext))
	fmt.Fprint(w, body)
}

func (s *Server) handleKeyGenerate(w http.ResponseWriter, r *http.Request) {
	if err := s.checkCSRF(r); err != nil {
		s.renderError(w, r, http.StatusForbidden, "Forbidden", "%s", err.Error())
		return
	}
	if session.KeysExist() {
		s.renderError(w, r, http.StatusConflict, "Signing keys already exist",
			"Refusing to overwrite the existing Ed25519 keypair. Remove %s to rotate.", session.KeysDir())
		return
	}
	pub, priv, err := session.GenerateKeys()
	if err != nil {
		s.renderError(w, r, http.StatusInternalServerError, "Key generation failed", "%s", err.Error())
		return
	}
	if err := session.SaveKeys(priv, pub); err != nil {
		s.renderError(w, r, http.StatusInternalServerError, "Could not save keys", "%s", err.Error())
		return
	}
	http.Redirect(w, r, "/sessions", http.StatusSeeOther)
}

// ─── execution API (SSE) ────────────────────────────────────────────────────

type runRequest struct {
	Tool    string   `json:"tool"`
	Args    []string `json:"args"`
	Session string   `json:"session"`
	Confirm bool     `json:"confirm"`
}

// handleRun executes a whitelisted tool and streams stdout/stderr to the
// client as Server-Sent Events. Execution always flows through pkg/runner so
// session audit logging is preserved.
func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if err := s.checkCSRF(r); err != nil {
		writeJSONError(w, http.StatusForbidden, err.Error())
		return
	}

	var req runRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// Whitelist the tool against the catalog — never trust the client.
	t, ok := catalog.Find(req.Tool)
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "unknown tool: "+req.Tool)
		return
	}

	av := runner.CheckAvailability(t.Binary)
	if !av.Available() {
		writeJSONError(w, http.StatusConflict,
			fmt.Sprintf("'%s' is not installed and no container runtime is available", t.Binary))
		return
	}

	// Destructive/root-needing tools require explicit confirmation.
	if t.NeedsRoot && !req.Confirm {
		writeJSONError(w, http.StatusBadRequest,
			t.Name+" may require elevated privileges; confirmation required")
		return
	}

	sessionID := strings.TrimSpace(req.Session)
	if sessionID != "" {
		sess, err := s.manager.Load(sessionID)
		if err != nil {
			writeJSONError(w, http.StatusNotFound, "session not found: "+sessionID)
			return
		}
		if sess.State != session.StateUnlocked {
			writeJSONError(w, http.StatusConflict,
				fmt.Sprintf("session %q is %s and read-only", sessionID, sess.State))
			return
		}
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	mode := "container"
	if av.Native {
		mode = "native"
	}
	writeSSE(w, flusher, "meta", map[string]any{
		"tool":      t.Name,
		"binary":    t.Binary,
		"mode":      mode,
		"needsRoot": t.NeedsRoot,
		"session":   sessionID,
	})

	outW := &sseWriter{w: w, f: flusher, event: "stdout"}
	errW := &sseWriter{w: w, f: flusher, event: "stderr"}

	opts := t.RunOpts(req.Args)
	opts.SessionID = sessionID
	err := runner.RunStream(r.Context(), opts, outW, errW)

	exitErr := ""
	if err != nil {
		exitErr = err.Error()
	}
	logged := ""
	if sessionID != "" {
		logged = sessionID
	}
	writeSSE(w, flusher, "exit", map[string]any{
		"code":   runner.ExitCode(err),
		"error":  exitErr,
		"logged": logged,
	})
}

// ─── helpers ────────────────────────────────────────────────────────────────

func (s *Server) buildCommandViews(sess *session.Session) []commandView {
	var pub ed25519.PublicKey
	if sess.Signed {
		pub, _ = sessionPublicKey(sess)
	}
	var out []commandView
	for _, c := range sess.Commands {
		v := commandView{CommandEntry: c}
		if c.Signature != "" {
			v.Sig.Present = true
			if pub != nil {
				v.Sig.Valid = session.Verify(pub, session.CommandSigningData(c), c.Signature)
			}
		}
		out = append(out, v)
	}
	return out
}

// sessionPublicKey prefers the key embedded in the session (self-contained
// chain of custody) and falls back to the global keypair.
func sessionPublicKey(sess *session.Session) (ed25519.PublicKey, error) {
	if sess.PublicKey != "" {
		return session.ParsePublicKey(sess.PublicKey)
	}
	return session.LoadPublicKey()
}

func (s *Server) verifySession(sess *session.Session) (state, msg string) {
	if !sess.Signed {
		return "unsigned", ""
	}
	pub, err := sessionPublicKey(sess)
	if err != nil {
		return "unavailable", "public key not found: " + err.Error()
	}
	if err := s.manager.VerifySession(sess, pub); err != nil {
		return "fail", err.Error()
	}
	return "ok", ""
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeSSE(w http.ResponseWriter, f http.Flusher, event string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	f.Flush()
}

// sseWriter streams writes as SSE events. It is safe for concurrent use
// because exec pipes stdout and stderr from separate goroutines.
type sseWriter struct {
	mu    sync.Mutex
	w     http.ResponseWriter
	f     http.Flusher
	event string
}

func (s *sseWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, _ := json.Marshal(string(p))
	fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", s.event, data)
	s.f.Flush()
	return len(p), nil
}
