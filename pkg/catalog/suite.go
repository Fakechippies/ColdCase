package catalog

import (
	"os"
	"path/filepath"
)

// DefaultSuitePath is the conventional, repo-relative location of the
// DidierStevensSuite scripts. It is only a fallback: use ResolveSuitePath to
// find the suite wherever it actually lives.
const DefaultSuitePath = "./DidierStevensSuite"

// suiteDirName is the directory name used by every discovery candidate.
const suiteDirName = "DidierStevensSuite"

// ResolveSuitePath returns the first existing DidierStevensSuite directory.
//
// Discovery order:
//  1. $COLDCASE_SUITE, an explicit override
//  2. ./DidierStevensSuite, then each parent directory (running from a
//     checkout, including subdirectories of the repo)
//  3. next to the executable, then one level above it (covers `bin/coldcase`
//     in a checkout and a suite installed beside an installed binary)
//  4. ~/.coldcase/DidierStevensSuite (created by `coldcase install`)
//  5. $XDG_DATA_HOME/coldcase/DidierStevensSuite
//
// When nothing exists, DefaultSuitePath is returned so error messages still
// point at the conventional location.
func ResolveSuitePath() string {
	if p := os.Getenv("COLDCASE_SUITE"); p != "" && isDir(p) {
		return abs(p)
	}

	if wd, err := os.Getwd(); err == nil {
		for _, p := range ancestorSuiteDirs(wd) {
			if isDir(p) {
				return abs(p)
			}
		}
	}

	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
		dir := filepath.Dir(exe)
		for _, p := range []string{
			filepath.Join(dir, suiteDirName),
			filepath.Join(filepath.Dir(dir), suiteDirName),
		} {
			if isDir(p) {
				return abs(p)
			}
		}
	}

	if home, err := os.UserHomeDir(); err == nil {
		if p := filepath.Join(home, ".coldcase", suiteDirName); isDir(p) {
			return abs(p)
		}
	}

	if data := os.Getenv("XDG_DATA_HOME"); data != "" {
		if p := filepath.Join(data, "coldcase", suiteDirName); isDir(p) {
			return abs(p)
		}
	}

	return DefaultSuitePath
}

// SuiteLinkPath is the per-user location `coldcase install` links the suite
// into so an installed binary can find it outside the source checkout.
func SuiteLinkPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".coldcase", suiteDirName), nil
}

// ancestorSuiteDirs lists <dir>/DidierStevensSuite for dir and every parent.
func ancestorSuiteDirs(dir string) []string {
	dir = abs(dir)
	var out []string
	for {
		out = append(out, filepath.Join(dir, suiteDirName))
		parent := filepath.Dir(dir)
		if parent == dir {
			return out
		}
		dir = parent
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func abs(path string) string {
	if p, err := filepath.Abs(path); err == nil {
		return p
	}
	return path
}
