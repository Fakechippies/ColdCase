package server

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"

	"coldcase/pkg/catalog"
	"coldcase/pkg/runner"
	"coldcase/pkg/session"
	"coldcase/pkg/tools"
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
	AvailSummary   availabilitySummary
	CategoryStats  []categoryStat
	StateStats     []stateStat
	Activity       []activityPoint
	TopTools       []toolUsage
	RecentCommands []commandDigest

	// tools
	Query        string
	Categories   []catalog.Category
	Availability map[string]availabilityView

	// tool detail
	Tool           catalog.Tool
	Avail          availabilityView
	SessionOptions []sessionOption
	Related        []catalog.Tool

	// status
	Rows           []statusRow
	AvailableCount int
	MissingCount   int

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

// ─── dashboard / availability aggregates ────────────────────────────────────

// availabilitySummary is the tool-level availability split rendered as a
// donut chart and legend.
type availabilitySummary struct {
	Total, Native, Container, Missing, Available int
	AvailablePct                                 int
	NativePct, ContainerPct, MissingPct          int
	Donut                                        []donutSegment
}

type donutSegment struct {
	Label  string
	Count  int
	Pct    int
	Class  string
	Dash   string
	Gap    string
	Offset string
}

// barSegment is one slice of a stacked coverage bar. X and W are expressed in
// SVG user units (0–100), pre-formatted to avoid template arithmetic.
type barSegment struct {
	Class string
	X     string
	W     string
}

type categoryStat struct {
	Name                     string
	Total, Native, Container int
	Missing, Available       int
	Segments                 []barSegment
}

type stateStat struct {
	Name  string
	Count int
	Pct   int
	Class string
}

type activityPoint struct {
	Label string
	Count int
	X, Y  int
	H     int
}

type toolUsage struct {
	Name  string
	Count int
	Pct   int
}

type commandDigest struct {
	Session    string
	Index      int
	Command    string
	Time       time.Time
	ExitCode   int
	DurationMS int64
}

const donutRadius = 54

var donutCircumference = 2 * math.Pi * donutRadius

func pctOf(part, total int) int {
	if total <= 0 {
		return 0
	}
	return int(math.Round(float64(part) * 100 / float64(total)))
}

func f2(v float64) string { return fmt.Sprintf("%.2f", v) }

// gatherAvailability walks the catalog once and returns both the overall
// availability split and per-category coverage. Binary lookups are cached so
// shared binaries are only probed once.
func gatherAvailability(cats []catalog.Category) (availabilitySummary, []categoryStat) {
	cache := map[string]bool{}
	containerReady := runner.DetectedRuntime() != ""
	nativeFor := func(binary string) bool {
		if v, ok := cache[binary]; ok {
			return v
		}
		v := tools.CheckToolInstalled(binary)
		cache[binary] = v
		return v
	}

	var total, nativeN, containerN, missingN int
	stats := make([]categoryStat, 0, len(cats))
	for _, c := range cats {
		cs := categoryStat{Name: c.Name, Total: len(c.Tools)}
		for _, t := range c.Tools {
			switch {
			case nativeFor(t.Binary):
				cs.Native++
				nativeN++
			case containerReady:
				cs.Container++
				containerN++
			default:
				cs.Missing++
				missingN++
			}
			total++
		}
		if cs.Total > 0 {
			cs.Available = cs.Native + cs.Container
			x := 0.0
			if cs.Native > 0 {
				w := float64(cs.Native) / float64(cs.Total) * 100
				if w < 1 {
					w = 1
				}
				cs.Segments = append(cs.Segments, barSegment{Class: "seg--native", X: f2(x), W: f2(w)})
				x += w
			}
			if cs.Container > 0 {
				w := float64(cs.Container) / float64(cs.Total) * 100
				if w < 1 {
					w = 1
				}
				cs.Segments = append(cs.Segments, barSegment{Class: "seg--container", X: f2(x), W: f2(w)})
				x += w
			}
			if cs.Missing > 0 {
				w := 100 - x
				if w < 1 {
					w = 1
				}
				cs.Segments = append(cs.Segments, barSegment{Class: "seg--missing", X: f2(x), W: f2(w)})
			}
		}
		stats = append(stats, cs)
	}

	return buildSummary(total, nativeN, containerN, missingN), stats
}

func buildSummary(total, nativeN, containerN, missingN int) availabilitySummary {
	s := availabilitySummary{
		Total:     total,
		Native:    nativeN,
		Container: containerN,
		Missing:   missingN,
	}
	s.Available = nativeN + containerN
	s.AvailablePct = pctOf(s.Available, total)
	s.NativePct = pctOf(nativeN, total)
	s.ContainerPct = pctOf(containerN, total)
	s.MissingPct = pctOf(missingN, total)

	denom := total
	if denom == 0 {
		denom = 1
	}
	offset := 0.0
	add := func(label string, count int, pct int, class string) {
		if count == 0 {
			return
		}
		dash := donutCircumference * float64(count) / float64(denom)
		s.Donut = append(s.Donut, donutSegment{
			Label:  label,
			Count:  count,
			Pct:    pct,
			Class:  class,
			Dash:   f2(dash),
			Gap:    f2(donutCircumference - dash),
			Offset: f2(-offset),
		})
		offset += dash
	}
	add("Native", nativeN, s.NativePct, "donut__seg--native")
	add("Container", containerN, s.ContainerPct, "donut__seg--container")
	add("Missing", missingN, s.MissingPct, "donut__seg--missing")
	return s
}

func (s *Server) loadSessions() []*session.Session {
	ids, err := s.manager.List()
	if err != nil {
		return nil
	}
	out := make([]*session.Session, 0, len(ids))
	for _, id := range ids {
		sess, err := s.manager.Load(id)
		if err != nil {
			continue
		}
		out = append(out, sess)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Created.After(out[j].Created) })
	return out
}

func summariesOf(sessions []*session.Session) []SessionSummary {
	out := make([]SessionSummary, 0, len(sessions))
	for _, sess := range sessions {
		out = append(out, SessionSummary{
			ID:           sess.ID,
			Investigator: sess.Investigator,
			State:        sess.State,
			Signed:       sess.Signed,
			CommandCount: len(sess.Commands),
			Created:      sess.Created,
		})
	}
	return out
}

// loadSummaries reads every session under ~/.coldcase/sessions, newest first.
func (s *Server) loadSummaries() []SessionSummary {
	return summariesOf(s.loadSessions())
}

func buildStateStats(sessions []*session.Session) []stateStat {
	total := len(sessions)
	counts := map[session.State]int{}
	for _, sess := range sessions {
		counts[sess.State]++
	}
	defs := []struct {
		state session.State
		name  string
		class string
	}{
		{session.StateUnlocked, "Unlocked", "unlocked"},
		{session.StateLocked, "Locked", "locked"},
		{session.StateSealed, "Sealed", "sealed"},
	}
	out := make([]stateStat, 0, len(defs))
	for _, d := range defs {
		n := counts[d.state]
		out = append(out, stateStat{Name: d.name, Count: n, Pct: pctOf(n, total), Class: d.class})
	}
	return out
}

func buildActivity(sessions []*session.Session) []activityPoint {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -6)
	counts := make([]int, 7)
	for _, sess := range sessions {
		for _, c := range sess.Commands {
			t := c.Timestamp.In(now.Location())
			if t.Before(start) {
				continue
			}
			day := int(t.Sub(start).Hours() / 24)
			if day >= 0 && day < len(counts) {
				counts[day]++
			}
		}
	}
	max := 1
	for _, n := range counts {
		if n > max {
			max = n
		}
	}
	out := make([]activityPoint, 0, len(counts))
	for i, n := range counts {
		h := 4 + int(math.Round(float64(n)/float64(max)*56))
		if n == 0 {
			h = 3
		}
		out = append(out, activityPoint{
			Label: start.AddDate(0, 0, i).Format("Mon"),
			Count: n,
			X:     8 + i*13,
			Y:     66 - h,
			H:     h,
		})
	}
	return out
}

func buildTopTools(sessions []*session.Session) []toolUsage {
	counts := map[string]int{}
	for _, sess := range sessions {
		for _, c := range sess.Commands {
			counts[c.Command]++
		}
	}
	type pair struct {
		name  string
		count int
	}
	pairs := make([]pair, 0, len(counts))
	for name, count := range counts {
		pairs = append(pairs, pair{name, count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].name < pairs[j].name
		}
		return pairs[i].count > pairs[j].count
	})
	if len(pairs) > 6 {
		pairs = pairs[:6]
	}
	max := 1
	if len(pairs) > 0 {
		max = pairs[0].count
	}
	out := make([]toolUsage, 0, len(pairs))
	for _, p := range pairs {
		pct := pctOf(p.count, max)
		if pct < 2 {
			pct = 2
		}
		out = append(out, toolUsage{Name: p.name, Count: p.count, Pct: pct})
	}
	return out
}

func buildRecentCommands(sessions []*session.Session, limit int) []commandDigest {
	var out []commandDigest
	for _, sess := range sessions {
		for _, c := range sess.Commands {
			out = append(out, commandDigest{
				Session:    sess.ID,
				Index:      c.Index,
				Command:    c.Command,
				Time:       c.Timestamp,
				ExitCode:   c.ExitCode,
				DurationMS: c.DurationMS,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.After(out[j].Time) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

// ─── helpers ────────────────────────────────────────────────────────────────

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

// loadSessionOptions returns session selector entries, newest first.
func (s *Server) loadSessionOptions(selected string) []sessionOption {
	var out []sessionOption
	for _, sum := range s.loadSummaries() {
		out = append(out, sessionOption{ID: sum.ID, State: sum.State, Selected: sum.ID == selected})
	}
	return out
}
