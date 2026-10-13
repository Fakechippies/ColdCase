// Package catalog is the single source of truth for the ColdCase tool
// catalog. Both the CLI (`coldcase list`, `coldcase check`) and the web API
// derive their view of the available tools from here so the two surfaces can
// never drift.
//
// Each tool records its category, description, whether it needs elevated
// privileges, and a RunOpts planner that describes exactly how to execute it
// through pkg/runner. The planner is the same code path the CLI uses, so
// execution semantics (native vs. container, path remapping, session logging)
// are shared.
package catalog

import (
	"sort"
	"sync"

	"coldcase/pkg/binwalk"
	"coldcase/pkg/carving"
	"coldcase/pkg/didier"
	"coldcase/pkg/exiftool"
	"coldcase/pkg/hashing"
	"coldcase/pkg/malware"
	"coldcase/pkg/mobile"
	"coldcase/pkg/network"
	"coldcase/pkg/runner"
	"coldcase/pkg/sleuthkit"
	"coldcase/pkg/steg"
	"coldcase/pkg/sysutils"
	"coldcase/pkg/timeline"
	vol3 "coldcase/pkg/volatility3"
	wintools "coldcase/pkg/windows"
)

// DefaultSuitePath is the directory holding the DidierStevensSuite Python
// scripts, relative to the working directory.
const DefaultSuitePath = "./DidierStevensSuite"

// Tool is a normalized catalog entry.
type Tool struct {
	// Name is the command name shown to the user (e.g. "pdfid" or "plaso parse").
	Name string
	// Description is a short one-line summary.
	Description string
	// Category is the display category this tool belongs to.
	Category string
	// NeedsRoot is true when the tool requires elevated/network capabilities.
	NeedsRoot bool
	// Binary is the underlying executable that will be invoked.
	Binary string
	// RunOpts builds the execution plan for a given argument vector.
	RunOpts func(args []string) runner.RunOpts
}

// Category groups tools under a display heading, preserving the canonical
// ordering used by the CLI.
type Category struct {
	Name  string
	Tools []Tool
}

// toolLike is the minimal interface every adapter satisfies. It mirrors the
// CLI's `tool` interface plus the RunOpts planner shared by all packages.
type toolLike interface {
	Name() string
	Description() string
	RunOpts(args []string) runner.RunOpts
}

// plan normalizes an adapter into a catalog Tool.
func plan(category string, t toolLike) Tool {
	binary := t.RunOpts(nil).Binary
	return Tool{
		Name:        t.Name(),
		Description: t.Description(),
		Category:    category,
		NeedsRoot:   t.RunOpts(nil).NeedsRoot,
		Binary:      binary,
		RunOpts:     t.RunOpts,
	}
}

// Categories returns the full catalog grouped and ordered exactly as the CLI
// has historically presented it.
func Categories() []Category {
	var cats []Category
	add := func(name string, tools []Tool) {
		cats = append(cats, Category{Name: name, Tools: tools})
	}

	// DidierStevens Suite
	var did []Tool
	for _, t := range didier.Tools(DefaultSuitePath) {
		did = append(did, plan("DidierStevens Suite", t))
	}
	add("DidierStevens Suite", did)

	// Network Forensics
	var net []Tool
	for _, t := range network.Tools() {
		net = append(net, plan("Network Forensics", t))
	}
	add("Network Forensics", net)

	// File Carving & Recovery
	var carve []Tool
	for _, t := range carving.Tools() {
		carve = append(carve, plan("File Carving & Recovery", t))
	}
	add("File Carving & Recovery", carve)

	// Malware & Pattern Matching
	var mal []Tool
	for _, t := range malware.Tools() {
		mal = append(mal, plan("Malware & Pattern Matching", t))
	}
	add("Malware & Pattern Matching", mal)

	// Hashing & Verification
	var hash []Tool
	for _, t := range hashing.Tools() {
		hash = append(hash, plan("Hashing & Verification", t))
	}
	add("Hashing & Verification", hash)

	// Timeline & Log Analysis (non-Plaso tools).
	var tl []Tool
	for _, t := range timeline.Tools() {
		switch t.Name() {
		case "log2timeline", "psort", "psteal":
			continue
		}
		tl = append(tl, plan("Timeline & Log Analysis", t))
	}
	add("Timeline & Log Analysis", tl)

	// Plaso command group (synthetic entries mapping onto the Plaso binaries).
	add("Plaso (log2timeline)", plasoTools())

	// Mobile Forensics
	var mob []Tool
	for _, t := range mobile.Tools() {
		mob = append(mob, plan("Mobile Forensics", t))
	}
	add("Mobile Forensics", mob)

	// Windows Artifacts
	var win []Tool
	for _, t := range wintools.Tools() {
		win = append(win, plan("Windows Artifacts", t))
	}
	add("Windows Artifacts", win)

	// Steganography & Media
	var stegT []Tool
	for _, t := range steg.Tools() {
		stegT = append(stegT, plan("Steganography & Media", t))
	}
	add("Steganography & Media", stegT)

	// System Utilities
	var sys []Tool
	for _, t := range sysutils.Tools() {
		sys = append(sys, plan("System Utilities", t))
	}
	add("System Utilities", sys)

	// Sleuth Kit
	var tsk []Tool
	for _, t := range sleuthkit.Tools() {
		tsk = append(tsk, plan("Sleuth Kit", t))
	}
	add("Sleuth Kit", tsk)

	// Volatility3
	var vol []Tool
	for _, t := range vol3.Tools() {
		vol = append(vol, plan("Volatility3 Memory Forensics", t))
	}
	add("Volatility3 Memory Forensics", vol)

	// General Tools
	add("General Tools", []Tool{
		plan("General Tools", exiftool.New()),
		plan("General Tools", binwalk.New()),
	})

	return cats
}

// plasoTools builds the synthetic Plaso command entries.
func plasoTools() []Tool {
	var byName = map[string]*timeline.TimelineTool{}
	for _, t := range timeline.Tools() {
		byName[t.Name()] = t
	}

	var out []Tool
	mk := func(display, source, desc string) {
		src := byName[source]
		if src == nil {
			return
		}
		baseOpts := src.RunOpts
		out = append(out, Tool{
			Name:        display,
			Description: desc,
			Category:    "Plaso (log2timeline)",
			Binary:      src.RunOpts(nil).Binary,
			RunOpts: func(args []string) runner.RunOpts {
				return baseOpts(args)
			},
		})
	}

	mk("plaso parse", "log2timeline", "Plaso — multi-source supertimeline generation")
	mk("plaso sort", "psort", "Plaso — sort and filter Plaso storage files")
	mk("plaso psteal", "psteal", "Plaso — extract and output timeline in one step")

	if l2t := byName["log2timeline"]; l2t != nil {
		base := l2t.RunOpts
		out = append(out, Tool{
			Name:        "plaso parsers",
			Description: "List all available plaso parsers",
			Category:    "Plaso (log2timeline)",
			Binary:      base(nil).Binary,
			RunOpts: func(args []string) runner.RunOpts {
				return base(append([]string{"--parsers", "list"}, args...))
			},
		})
	}

	return out
}

// All returns every tool as a flat list, in category order.
func All() []Tool {
	var out []Tool
	for _, c := range Categories() {
		out = append(out, c.Tools...)
	}
	return out
}

var (
	indexOnce sync.Once
	index     map[string]Tool
)

// Find looks up a tool by command name. The second result is false when the
// name is not part of the known catalog; callers must treat that as a
// rejection (the web API whitelists tool names against this).
func Find(name string) (Tool, bool) {
	indexOnce.Do(func() {
		index = make(map[string]Tool)
		for _, t := range All() {
			index[t.Name] = t
		}
	})
	t, ok := index[name]
	return t, ok
}

// Names returns the sorted list of all known tool names. Useful for
// diagnostics and tests.
func Names() []string {
	names := make([]string, 0, len(All()))
	for _, t := range All() {
		names = append(names, t.Name)
	}
	sort.Strings(names)
	return names
}
