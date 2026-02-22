// Package timeline integrates timeline generation and log analysis tools.
package timeline

import (
	"coldcase/pkg/runner"
)

// TimelineTool wraps a timeline or log analysis binary.
type TimelineTool struct {
	name string
	bin  string
	desc string
}

func (t *TimelineTool) Name() string        { return t.name }
func (t *TimelineTool) Description() string { return t.desc }
func (t *TimelineTool) Run(args []string) error {
	bin := t.bin
	if bin == "" {
		bin = t.name
	}
	return runner.Run(runner.RunOpts{Binary: bin, Args: args})
}

// Tools returns all timeline and log analysis tools.
func Tools() []*TimelineTool {
	return []*TimelineTool{
		{"log2timeline", "log2timeline", "Plaso — multi-source supertimeline generation"},
		{"psteal", "psteal", "Plaso — extract and output timeline in one step"},
		{"psort", "psort", "Plaso — sort and filter Plaso storage files"},
		{"hayabusa", "", "Sigma-based threat hunting and fast timeline from Windows EVTX logs"},
		{"evtx_dump", "", "Parse and convert Windows EVTX event log files"},
		{"timeliner", "", "mactime bodyfile reader and timeline generator"},
		{"chainsaw", "", "Rapid Windows event log analysis with Sigma and built-in rules"},
	}
}
