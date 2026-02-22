package session

import (
	"time"
)

type State string

const (
	StateUnlocked State = "unlocked"
	StateLocked   State = "locked"
	StateSealed   State = "sealed"
)

type Session struct {
	ID           string         `json:"id"`
	Investigator string         `json:"investigator"`
	Email        string         `json:"email"`
	Created      time.Time      `json:"created"`
	State        State          `json:"state"`
	Encrypted    bool           `json:"encrypted"`
	Signed       bool           `json:"signed"`
	PublicKey    string         `json:"public_key,omitempty"`
	Commands     []CommandEntry `json:"commands"`
	Evidence     []EvidenceFile `json:"evidence_files"`
	Signature    string         `json:"signature,omitempty"` // Final session signature
	SealedAt     *time.Time     `json:"sealed_at,omitempty"`
}

type CommandEntry struct {
	Index            int            `json:"index"`
	Timestamp        time.Time      `json:"timestamp"`
	Command          string         `json:"command"` // Base command e.g. "pdfid"
	FullCommand      string         `json:"full_command"`
	Args             []string       `json:"args"`
	InputFiles       []FileMetadata `json:"input_files"`
	ToolPath         string         `json:"tool_path"`
	ToolVersion      string         `json:"tool_version"`
	ExitCode         int            `json:"exit_code"`
	DurationMS       int64          `json:"duration_ms"`
	OutputPreview    string         `json:"output_preview"`
	OutputFile       string         `json:"output_file"`
	WorkingDirectory string         `json:"working_directory"`
	Signature        string         `json:"signature,omitempty"`
}

type FileMetadata struct {
	Path     string    `json:"path"`
	SHA256   string    `json:"sha256"`
	MD5      string    `json:"md5"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

type EvidenceFile struct {
	OriginalPath string    `json:"original_path"`
	SHA256       string    `json:"sha256"`
	CapturedAt   time.Time `json:"captured_at"`
	Size         int64     `json:"file_size"`
}
