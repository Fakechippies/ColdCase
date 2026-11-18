package session

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// SessionIDPattern is the allow-list for session identifiers. It doubles as a
// path-safety guard: everything session IDs are used in filesystem paths, so
// they must never contain separators or traversal characters.
var SessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// ValidSessionID reports whether id is safe to use in paths and commands.
// "." and ".." are explicitly rejected: they are valid under the character
// set but traverse the sessions directory.
func ValidSessionID(id string) bool {
	if id == "." || id == ".." {
		return false
	}
	return SessionIDPattern.MatchString(id)
}

var (
	baseDir string
)

func init() {
	home, _ := os.UserHomeDir()
	baseDir = filepath.Join(home, ".coldcase")
}

// Manager handles session persistence and environment integration
type Manager struct {
	sessionsDir string
}

func NewManager() *Manager {
	dir := filepath.Join(baseDir, "sessions")
	_ = os.MkdirAll(dir, 0700)
	return &Manager{sessionsDir: dir}
}

func (m *Manager) Create(id string, investigator, email string, sign, encrypt bool) (*Session, error) {
	if !ValidSessionID(id) {
		return nil, fmt.Errorf("invalid session id %q: use 1-64 characters of [A-Za-z0-9._-]", id)
	}
	sDir := filepath.Join(m.sessionsDir, id)
	if _, err := os.Stat(sDir); err == nil {
		return nil, fmt.Errorf("session '%s' already exists", id)
	}

	if err := os.MkdirAll(filepath.Join(sDir, "outputs"), 0700); err != nil {
		return nil, err
	}

	s := &Session{
		ID:           id,
		Investigator: investigator,
		Email:        email,
		Created:      time.Now(),
		State:        StateUnlocked,
		Signed:       sign,
		Encrypted:    encrypt,
		Commands:     []CommandEntry{},
		Evidence:     []EvidenceFile{},
	}

	// Embed the active public key so the session remains independently
	// verifiable even if the global keypair is later rotated.
	if sign {
		if pub, err := LoadPublicKey(); err == nil {
			s.PublicKey = hex.EncodeToString(pub)
		}
	}

	if err := m.Save(s); err != nil {
		return nil, err
	}

	return s, nil
}

func (m *Manager) Load(id string) (*Session, error) {
	if !ValidSessionID(id) {
		return nil, fmt.Errorf("invalid session id %q", id)
	}
	path := filepath.Join(m.sessionsDir, id, "session.json")
	data, err := os.ReadFile(path) //nolint:gosec // id is validated against SessionIDPattern above
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (m *Manager) Save(s *Session) error {
	if !ValidSessionID(s.ID) {
		return fmt.Errorf("invalid session id %q", s.ID)
	}
	path := filepath.Join(m.sessionsDir, s.ID, "session.json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.sessionsDir)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, e := range entries {
		if e.IsDir() {
			results = append(results, e.Name())
		}
	}
	return results, nil
}

// Lock transitions a session to the read-only "locked" state.
func (m *Manager) Lock(s *Session) error {
	s.State = StateLocked
	return m.Save(s)
}

// Unlock returns a locked session to the "unlocked" state so tools may run.
func (m *Manager) Unlock(s *Session) error {
	s.State = StateUnlocked
	s.SealedAt = nil
	return m.Save(s)
}

// Seal permanently closes a session. If signing is enabled and a private key
// is present, a final session signature is produced over the sealed summary.
func (m *Manager) Seal(s *Session) error {
	now := time.Now()
	s.State = StateSealed
	s.SealedAt = &now
	if s.Signed {
		if priv, err := LoadPrivateKey(); err == nil {
			s.Signature = Sign(priv, SessionSigningData(s))
		}
	}
	return m.Save(s)
}

// GetActiveSessionID returns the session ID from the environment
func GetActiveSessionID() string {
	return os.Getenv("COLDCASE_SESSION_ID")
}
