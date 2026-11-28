package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolveSuitePathFromAncestor covers the common case: the binary runs
// from a subdirectory of the checkout, so the suite is found by walking up.
func TestResolveSuitePathFromAncestor(t *testing.T) {
	root := t.TempDir()
	suite := filepath.Join(root, suiteDirName)
	if err := os.MkdirAll(suite, 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "cmd", "coldcase")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COLDCASE_SUITE", "")
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Chdir(nested)

	if got, want := ResolveSuitePath(), abs(suite); got != want {
		t.Fatalf("ResolveSuitePath() = %q, want %q", got, want)
	}
}

// TestResolveSuitePathEnvOverride verifies the explicit override wins even
// when another suite exists closer to the working directory.
func TestResolveSuitePathEnvOverride(t *testing.T) {
	wd := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wd, suiteDirName), 0o755); err != nil {
		t.Fatal(err)
	}
	override := filepath.Join(t.TempDir(), "custom-suite")
	if err := os.MkdirAll(override, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COLDCASE_SUITE", override)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Chdir(wd)

	if got, want := ResolveSuitePath(), abs(override); got != want {
		t.Fatalf("ResolveSuitePath() = %q, want %q", got, want)
	}
}

// TestResolveSuitePathHomeLink covers the installed-binary case: no suite in
// the working tree, but `coldcase install` linked one into ~/.coldcase.
func TestResolveSuitePathHomeLink(t *testing.T) {
	home := t.TempDir()
	linked := filepath.Join(home, ".coldcase", suiteDirName)
	if err := os.MkdirAll(linked, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COLDCASE_SUITE", "")
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", home)
	t.Chdir(t.TempDir())

	if got, want := ResolveSuitePath(), abs(linked); got != want {
		t.Fatalf("ResolveSuitePath() = %q, want %q", got, want)
	}
}

// TestResolveSuitePathFallback verifies the conventional relative path is
// returned when nothing can be found, so diagnostics stay meaningful.
func TestResolveSuitePathFallback(t *testing.T) {
	t.Setenv("COLDCASE_SUITE", "")
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Chdir(t.TempDir())

	if got := ResolveSuitePath(); got != DefaultSuitePath {
		t.Fatalf("ResolveSuitePath() = %q, want fallback %q", got, DefaultSuitePath)
	}
}
