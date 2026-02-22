// Package network integrates network forensics tools into ColdCase.
// All tools are run natively if available, otherwise via the container runtime.
package network

import (
	"coldcase/pkg/runner"
)

// NetworkTool wraps a network forensics binary.
type NetworkTool struct {
	name      string
	desc      string
	needsRoot bool // tools like tcpdump need NET_ADMIN
}

func (n *NetworkTool) Name() string        { return n.name }
func (n *NetworkTool) Description() string { return n.desc }
func (n *NetworkTool) Run(args []string) error {
	return runner.Run(runner.RunOpts{
		Binary:    n.name,
		Args:      args,
		NeedsRoot: n.needsRoot,
	})
}

// Tools returns all network forensics tools.
func Tools() []*NetworkTool {
	return []*NetworkTool{
		{"tshark", "CLI Wireshark — PCAP dissection and analysis", false},
		{"tcpdump", "Packet capture and filtering", true},
		{"zeek", "Network security monitoring and traffic analysis", true},
		{"ngrep", "grep over network packets", true},
		{"tcpflow", "TCP flow reconstruction from PCAPs", false},
		{"pcapfix", "Repair corrupted PCAP/PCAPng files", false},
		{"tcpreplay", "Replay PCAP files onto a network interface", true},
		{"tcpstat", "Network interface statistics from PCAP files", false},
		{"argus", "Network flow analysis and auditing", true},
		{"p0f", "Passive OS fingerprinting from PCAP", true},
		{"networkminer", "Network Forensic Analysis Tool (CLI)", false},
	}
}
