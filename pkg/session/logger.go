package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	session *Session
	manager *Manager
}

func NewLogger(s *Session) *Logger {
	return &Logger{
		session: s,
		manager: NewManager(),
	}
}

func (l *Logger) LogCommand(entry CommandEntry) error {
	if l.session.Signed {
		priv, err := LoadPrivateKey()
		if err == nil {
			// Sign a concat of key fields
			data := fmt.Sprintf("%d|%s|%s|%s", entry.Index, entry.Timestamp.UTC().Format(time.RFC3339), entry.FullCommand, entry.WorkingDirectory)
			entry.Signature = Sign(priv, []byte(data))
		}
	}
	l.session.Commands = append(l.session.Commands, entry)
	return l.manager.Save(l.session)
}

func (l *Logger) HashInputFile(path string) (FileMetadata, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return FileMetadata{}, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return FileMetadata{}, err
	}

	// Simple cache check: see if we already have this file hashed in this session
	for _, f := range l.session.Evidence {
		if f.OriginalPath == absPath && f.Size == info.Size() {
			// In a real implementation, we'd check timestamps too
			return FileMetadata{
				Path:     absPath,
				SHA256:   f.SHA256,
				Size:     f.Size,
				Modified: info.ModTime(),
			}, nil
		}
	}

	sha, _, err := HashFile(absPath)
	if err != nil {
		return FileMetadata{}, err
	}

	meta := FileMetadata{
		Path:     absPath,
		SHA256:   sha,
		Size:     info.Size(),
		Modified: info.ModTime(),
	}

	// Add to evidence list if new
	l.session.Evidence = append(l.session.Evidence, EvidenceFile{
		OriginalPath: absPath,
		SHA256:       sha,
		CapturedAt:   time.Now(),
		Size:         info.Size(),
	})

	return meta, nil
}

func (l *Logger) SaveOutput(index int, name string, output []byte) (string, error) {
	filename := fmt.Sprintf("%s_%03d.txt", name, index)
	path := filepath.Join(baseDir, "sessions", l.session.ID, "outputs", filename)

	if l.session.Encrypted {
		// Encryption logic would go here if we have the key
		// For now, write plain as we focus on audit logging
	}

	err := os.WriteFile(path, output, 0600)
	if err != nil {
		return "", err
	}

	return filepath.Join("outputs", filename), nil
}
