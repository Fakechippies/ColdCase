package session

import (
	"testing"
)

func TestValidSessionID(t *testing.T) {
	valid := []string{"case-001", "inv_2", "a.b", "A9", strings_n(63, "x")}
	for _, id := range valid {
		if !ValidSessionID(id) {
			t.Fatalf("ValidSessionID(%q) = false, want true", id)
		}
	}

	invalid := []string{
		"",                 // empty
		strings_n(65, "x"), // too long
		"../escape",        // traversal
		"a/b",              // path separator
		"a\\b",             // windows separator
		"sp ace",           // space
		"semi;colon",       // shell metacharacter
		"dot.t./",          // trailing separator
		"ünïcode",          // non-ascii
		"..",               // traversal only
	}
	for _, id := range invalid {
		if ValidSessionID(id) {
			t.Fatalf("ValidSessionID(%q) = true, want false", id)
		}
	}
}

// strings_n builds a string of n copies of s (strconv is overkill here).
func strings_n(n int, s string) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

// TestManagerRejectsUnsafeIDs proves the model layer refuses identifiers that
// could escape the sessions directory, even when callers forget to validate.
func TestManagerRejectsUnsafeIDs(t *testing.T) {
	// Redirect the sessions root to a temp dir: tests must never touch the
	// real ~/.coldcase state.
	baseDir = t.TempDir()
	m := NewManager()

	if _, err := m.Load("../escape"); err == nil {
		t.Fatal("Load accepted a traversal id")
	}
	if _, err := m.Load("a/b"); err == nil {
		t.Fatal("Load accepted an id with a path separator")
	}
	if _, err := m.Create("../escape", "", "", false, false); err == nil {
		t.Fatal("Create accepted a traversal id")
	}

	// A valid id still round-trips.
	s, err := m.Create("id-test-001", "tester", "", false, false)
	if err != nil {
		t.Fatalf("Create valid id: %v", err)
	}
	if err := m.Save(s); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := m.Load("id-test-001"); err != nil {
		t.Fatalf("Load: %v", err)
	}
}
