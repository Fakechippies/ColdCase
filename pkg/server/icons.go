package server

import (
	"fmt"
	"html/template"
	"strings"
)

// svg wraps trusted, static icon path data in a consistent 24×24 stroke icon.
// Only hard-coded strings from this file are ever returned as template.HTML.
func svg(body string, size int) template.HTML {
	if size <= 0 {
		size = 18
	}
	return template.HTML(fmt.Sprintf(
		`<svg viewBox="0 0 24 24" width="%d" height="%d" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">%s</svg>`,
		size, size, body))
}

const (
	iconFingerprint = `<path d="M12 10a2 2 0 0 0-2 2c0 1.02-.1 2.51-.26 4"/><path d="M14 13.12c0 2.38 0 6.38-1 8.88"/><path d="M17.29 21.02c.12-.6.43-2.3.5-3.02"/><path d="M2 12a10 10 0 0 1 18-6"/><path d="M2 16h.01"/><path d="M21.8 16c.2-2 .131-5.354 0-6"/><path d="M5 19.5C5.5 18 6 15 6 12a6 6 0 0 1 .34-2"/><path d="M8.65 22c.21-.66.45-1.32.57-2"/><path d="M9 6.8a6 6 0 0 1 9 5.2v2"/>`
	iconFile        = `<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M8 13h8"/><path d="M8 17h6"/>`
	iconImage       = `<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-4.35-4.35a2 2 0 0 0-2.83 0L4 21"/>`
	iconMail        = `<rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/>`
	iconNetwork     = `<path d="M5 12.55a11 11 0 0 1 14 0"/><path d="M8.5 16.05a6 6 0 0 1 7 0"/><path d="M12 20h.01"/><path d="M2 9a16 16 0 0 1 20 0"/>`
	iconMemory      = `<rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/><path d="M9 2v2"/><path d="M15 2v2"/><path d="M9 20v2"/><path d="M15 20v2"/><path d="M2 9h2"/><path d="M2 15h2"/><path d="M20 9h2"/><path d="M20 15h2"/>`
	iconClock       = `<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>`
	iconScissors    = `<path d="M14.7 6.3a4 4 0 0 0-5.4 5.4L3 18l3 3 6.3-6.3a4 4 0 0 0 5.4-5.4l-2.1 2.1-2.9-2.9 2-2.2Z"/>`
	iconShield      = `<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/><path d="M12 8v4"/><path d="M12 16h.01"/>`
	iconShieldCheck = `<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/><path d="m9 12 2 2 4-4"/>`
	iconMobile      = `<rect x="6" y="2" width="12" height="20" rx="2"/><path d="M12 18h.01"/>`
	iconDisk        = `<path d="M22 12H2"/><path d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11Z"/><path d="M6 16h.01"/><path d="M10 16h.01"/>`
	iconCode        = `<path d="m16 18 6-6-6-6"/><path d="m8 6-6 6 6 6"/><path d="m14.5 4-5 16"/>`
	iconSearch      = `<circle cx="11" cy="11" r="7"/><path d="m21 21-4.3-4.3"/>`
	iconBinary      = `<rect x="3" y="3" width="18" height="18" rx="2"/><path d="M8 8h.01"/><path d="M12 8h.01"/><path d="M16 8h.01"/><path d="M8 12h.01"/><path d="M12 12h.01"/><path d="M16 12h.01"/><path d="M8 16h.01"/><path d="M12 16h.01"/><path d="M16 16h.01"/>`
	iconMonitor     = `<rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8"/><path d="M12 17v4"/>`
	iconPackage     = `<path d="m7.5 4.27 9 5.15"/><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/>`
	iconTerminal    = `<rect x="2" y="4" width="20" height="16" rx="2"/><path d="m6 9 3 3-3 3"/><path d="M12 15h5"/>`
	iconActivity    = `<path d="M22 12h-4l-3 9L9 3l-3 9H2"/>`
)

// categoryIcon returns an icon for a catalog category.
func categoryIcon(name string) template.HTML {
	switch name {
	case "DidierStevens Suite":
		return svg(iconFile, 18)
	case "Network Forensics":
		return svg(iconNetwork, 18)
	case "File Carving & Recovery":
		return svg(iconPackage, 18)
	case "Malware & Pattern Matching":
		return svg(iconShield, 18)
	case "Hashing & Verification":
		return svg(iconFingerprint, 18)
	case "Timeline & Log Analysis", "Plaso (log2timeline)":
		return svg(iconClock, 18)
	case "Mobile Forensics":
		return svg(iconMobile, 18)
	case "Windows Artifacts":
		return svg(iconMonitor, 18)
	case "Steganography & Media":
		return svg(iconImage, 18)
	case "System Utilities":
		return svg(iconTerminal, 18)
	case "Sleuth Kit":
		return svg(iconDisk, 18)
	case "Volatility3 Memory Forensics":
		return svg(iconMemory, 18)
	case "General Tools":
		return svg(iconScissors, 18)
	default:
		return svg(iconPackage, 18)
	}
}

// toolIcon picks an icon from the tool's binary/name. It is intentionally
// heuristic: the catalog has 100+ tools and no per-tool icon metadata yet.
func toolIcon(binary, name string) template.HTML {
	key := strings.ToLower(binary + " " + name)
	has := func(terms ...string) bool {
		for _, t := range terms {
			if strings.Contains(key, t) {
				return true
			}
		}
		return false
	}
	switch {
	case has("hash", "md5", "sha", "ssdeep", "verify", "checksum"):
		return svg(iconFingerprint, 18)
	case has("pdf", "1768"):
		return svg(iconFile, 18)
	case has("eml", "mail", "smtp", "mbox"):
		return svg(iconMail, 18)
	case has("exif", "image", "jpg", "jpeg", "png", "gif", "steg", "zsteg", "steghide", "outguess", "jphide"):
		return svg(iconImage, 18)
	case has("tshark", "tcpdump", "network", "pcap", "dns", "http", "tcp", "udp", "traffic",
		"zeek", "ngrep", "tcpflow", "tcpreplay", "pcapfix", "topstat", "url"):
		return svg(iconNetwork, 18)
	case has("volatility", "memory", "windows.", "linux.", "mac."):
		return svg(iconMemory, 18)
	case has("log2timeline", "plaso", "psort", "timeline", "psteal", "hayabusa", "chainsaw"):
		return svg(iconClock, 18)
	case has("evtx", "registry", "regripper", "prefetch", "amcache", "shimcache", "windows"):
		return svg(iconMonitor, 18)
	case has("carv", "foremost", "scalpel", "photorec", "binwalk", "bulk"):
		return svg(iconScissors, 18)
	case has("malware", "yara", "clam", "virus", "cape", "amsiscan", "pecheck", "pe file", "oledump"):
		return svg(iconShield, 18)
	case has("mobile", "android", "ios", "apk", "adb", "plist"):
		return svg(iconMobile, 18)
	case has("sleuth", "fls", "icat", "mmls", "disk", "tsk", "fsstat", "blkls", "blkcat",
		"istat", "hfind", "mactime"):
		return svg(iconDisk, 18)
	case has("base64", "byte", "cut-bytes", "script", "code", "strings", "decode", "encode"):
		return svg(iconCode, 18)
	case has("find-file", "find_file", "search", "grep"):
		return svg(iconSearch, 18)
	case has("binary", "hex", "dump"):
		return svg(iconBinary, 18)
	default:
		return svg(iconScissors, 18)
	}
}

// fmtDuration renders a millisecond duration compactly.
func fmtDuration(ms int64) string {
	switch {
	case ms < 1000:
		return fmt.Sprintf("%dms", ms)
	case ms < 60000:
		return fmt.Sprintf("%.1fs", float64(ms)/1000)
	default:
		return fmt.Sprintf("%dm%02ds", ms/60000, (ms%60000)/1000)
	}
}
