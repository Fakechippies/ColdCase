// Package mobile integrates mobile device forensics tools for iOS and Android.
package mobile

import (
	"coldcase/pkg/runner"
)

// MobileTool wraps a mobile forensics binary.
type MobileTool struct {
	name string
	bin  string
	desc string
}

func (m *MobileTool) Name() string        { return m.name }
func (m *MobileTool) Description() string { return m.desc }
func (m *MobileTool) RunOpts(args []string) runner.RunOpts {
	bin := m.bin
	if bin == "" {
		bin = m.name
	}
	return runner.RunOpts{Binary: bin, Args: args}
}
func (m *MobileTool) Run(args []string) error {
	return runner.Run(m.RunOpts(args))
}

// Tools returns all mobile forensics tools.
func Tools() []*MobileTool {
	return []*MobileTool{
		{"aleapp", "aleapp", "Android Logs Events And Protobuf Parser (ALEAPP)"},
		{"ileapp", "ileapp", "iOS Logs Events And Plists Parser (iLEAPP)"},
		{"adb", "", "Android Debug Bridge — device interaction and artifact extraction"},
		{"ideviceinfo", "", "iOS device information and artifact extraction (libimobiledevice)"},
	}
}
