package server

import (
	"net/http"
	"sort"
	"time"

	"coldcase/pkg/catalog"
	"coldcase/pkg/runner"
	"coldcase/pkg/session"
)

// viewData is the union of every field the templates may reference. HTML
// templates error on missing struct fields, so all pages share one permissive
// struct; unused fields stay at their zero value.
type viewData struct {
	// layout
	Title          string
	Subtitle       string
	Nav            string
	CSRF           string
	ContainerReady bool
	Runtime        string
	Image          string
	KeysReady      bool
	ToolCount      int
	CategoryCount  int
	Flash          string
	FlashKind      string

	// dashboard
	NativeCount    int
	ContainerCount int
	SessionCount   int
	RecentSessions []SessionSummary

	// tools
	Query        string
	Categories   []catalog.Category
	Availability map[string]availabilityView

	// tool detail
	Tool           catalog.Tool
	Avail          availabilityView
	SessionOptions []sessionOption

	// status
	Rows           []statusRow
	AvailableCount int

	// sessions list
	Sessions []SessionSummary

	// session detail
	Session     *session.Session
	Commands    []commandView
	VerifyState string
	VerifyMsg   string

	// error
	ErrorCode   string
	ErrorTitle  string
	ErrorDetail string
}

// availabilityView is a template-friendly tool availability summary.
type availabilityView struct {
	Native    bool
	Container bool
}

// SessionSummary is a compact session row for list views.
type SessionSummary struct {
	ID           string
	Investigator string
	State        session.State
	Signed       bool
	CommandCount int
	Created      time.Time
}

// sessionOption is an entry in the run form's session selector.
type sessionOption struct {
	ID       string
	State    session.State
	Selected bool
}

// sigView describes whether a command signature is present and valid.
type sigView struct {
	Present bool
	Valid   bool
}

// commandView augments a command entry with its verification result.
type commandView struct {
	session.CommandEntry
	Sig sigView
}

// statusRow is one row of the availability table.
type statusRow struct {
	Name           string
	Category       string
	Binary         string
	Native         bool
	ContainerReady bool
	Available      bool
}

// CSRFToken returns the active double-submit token. It prefers the token
// resolved by middleware (which also serves the first, cookie-less request).
func CSRFToken(r *http.Request) string {
	if v, ok := r.Context().Value(csrfContextKey{}).(string); ok && v != "" {
		return v
	}
	if c, err := r.Cookie(csrfCookie); err == nil {
		return c.Value
	}
	return ""
}

// newData seeds the fields shared by every page.
func (s *Server) newData(title, nav string, r *http.Request) *viewData {
	rt := runner.DetectedRuntime()
	return &viewData{
		Title:          title,
		Nav:            nav,
		CSRF:           CSRFToken(r),
		ContainerReady: rt != "",
		Runtime:        rt,
		Image:          runner.ImageName(),
		KeysReady:      session.KeysExist(),
		ToolCount:      s.toolCount,
		CategoryCount:  s.categoryCount,
	}
}

// loadSummaries reads every session under ~/.coldcase/sessions, newest first.
func (s *Server) loadSummaries() []SessionSummary {
	ids, err := s.manager.List()
	if err != nil {
		return nil
	}
	var out []SessionSummary
	for _, id := range ids {
		sess, err := s.manager.Load(id)
		if err != nil {
			continue
		}
		out = append(out, SessionSummary{
			ID:           sess.ID,
			Investigator: sess.Investigator,
			State:        sess.State,
			Signed:       sess.Signed,
			CommandCount: len(sess.Commands),
			Created:      sess.Created,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Created.After(out[j].Created) })
	return out
}

// loadSessionOptions returns session selector entries, newest first.
func (s *Server) loadSessionOptions(selected string) []sessionOption {
	var out []sessionOption
	for _, sum := range s.loadSummaries() {
		out = append(out, sessionOption{ID: sum.ID, State: sum.State, Selected: sum.ID == selected})
	}
	return out
}
