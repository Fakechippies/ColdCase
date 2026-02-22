// Package steg integrates steganography detection and media forensics tools.
package steg

import (
	"coldcase/pkg/runner"
)

// StegTool wraps a steganography or media analysis binary.
type StegTool struct {
	name string
	bin  string
	desc string
}

func (s *StegTool) Name() string        { return s.name }
func (s *StegTool) Description() string { return s.desc }
func (s *StegTool) Run(args []string) error {
	bin := s.bin
	if bin == "" {
		bin = s.name
	}
	return runner.Run(runner.RunOpts{Binary: bin, Args: args})
}

// Tools returns all steganography and media tools.
func Tools() []*StegTool {
	return []*StegTool{
		{"steghide", "", "Embed and extract hidden data from image and audio files"},
		{"zsteg", "", "Detect steganographic content in PNG and BMP images"},
		{"wavsteg", "wavsteg", "Hide and detect data in WAV audio files"},
		{"mediainfo", "", "Display technical and tag metadata for video and audio files"},
		{"stegdetect", "", "Automated tool for detecting steganographic content in JPEGs"},
	}
}
