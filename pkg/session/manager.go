package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	baseDir string
	mu      sync.Mutex
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

	if err := m.Save(s); err != nil {
		return nil, err
	}

	return s, nil
}

func (m *Manager) Load(id string) (*Session, error) {
	path := filepath.Join(m.sessionsDir, id, "session.json")
	data, err := os.ReadFile(path)
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

// GetActiveSessionID returns the session ID from the environment
func GetActiveSessionID() string {
	return os.Getenv("COLDCASE_SESSION_ID")
}
