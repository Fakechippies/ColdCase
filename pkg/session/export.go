package session

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExportJSON returns the session data as a formatted JSON string
func ExportJSON(s *Session) (string, error) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ExportMarkdown returns the session as a technical markdown report
func ExportMarkdown(s *Session) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Forensic Investigation Report: %s\n\n", s.ID))
	sb.WriteString(fmt.Sprintf("- **Investigator**: %s (%s)\n", s.Investigator, s.Email))
	sb.WriteString(fmt.Sprintf("- **Created**: %s\n", s.Created.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("- **State**: %s\n", s.State))
	if s.SealedAt != nil {
		sb.WriteString(fmt.Sprintf("- **Sealed**: %s\n", s.SealedAt.Format("2006-01-02 15:04:05")))
	}
	sb.WriteString(fmt.Sprintf("- **Signed**: %v\n", s.Signed))
	sb.WriteString(fmt.Sprintf("- **Encrypted**: %v\n", s.Encrypted))
	sb.WriteString("\n## Command History\n\n")

	for _, cmd := range s.Commands {
		sb.WriteString(fmt.Sprintf("### [%d] %s\n", cmd.Index, cmd.Command))
		sb.WriteString(fmt.Sprintf("- **Time**: %s\n", cmd.Timestamp.Format("2006-01-02 15:04:05.000")))
		sb.WriteString(fmt.Sprintf("- **Full Command**: `%s`\n", cmd.FullCommand))
		sb.WriteString(fmt.Sprintf("- **Working Dir**: `%s`\n", cmd.WorkingDirectory))
		sb.WriteString(fmt.Sprintf("- **Duration**: %dms\n", cmd.DurationMS))
		if cmd.Signature != "" {
			sb.WriteString(fmt.Sprintf("- **Signature**: `%s`\n", cmd.Signature))
		}
		sb.WriteString("\n#### Output Preview\n```\n")
		sb.WriteString(cmd.OutputPreview)
		sb.WriteString("\n```\n\n")
	}

	sb.WriteString("## Evidence Integrity\n\n")
	sb.WriteString("| Path | SHA256 | Captured |\n")
	sb.WriteString("|------|--------|----------|\n")
	for _, f := range s.Evidence {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", f.OriginalPath, f.SHA256, f.CapturedAt.Format("2006-01-02 15:04:05")))
	}

	return sb.String(), nil
}

// ExportHTML returns a standalone HTML report
func ExportHTML(s *Session) (string, error) {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html><html><head><title>ColdCase Report - " + s.ID + "</title>")
	sb.WriteString("<style>body{font-family:sans-serif;line-height:1.6;color:#333;max-width:900px;margin:auto;padding:20px}")
	sb.WriteString("h1,h2,h3{color:#222} .cmd{border:1px solid #ddd;padding:15px;margin-bottom:20px;border-radius:4px;background:#f9f9f9}")
	sb.WriteString("pre{background:#eee;padding:10px;overflow-x:auto} .meta{font-size:0.9em;color:#666}")
	sb.WriteString("table{width:100%;border-collapse:collapse} th,td{border:1px solid #ddd;padding:8px;text-align:left} th{background:#eee}</style></head><body>")

	sb.WriteString("<h1>Forensic Investigation Report</h1>")
	sb.WriteString("<div class='meta'><p>Session ID: " + s.ID + "<br>")
	sb.WriteString("Investigator: " + s.Investigator + " (" + s.Email + ")<br>")
	sb.WriteString("Created: " + s.Created.Format("2006-01-02 15:04:05") + "<br>")
	sb.WriteString("State: " + string(s.State) + "</p></div>")

	sb.WriteString("<h2>Command History</h2>")
	for _, cmd := range s.Commands {
		sb.WriteString("<div class='cmd'>")
		sb.WriteString("<h3>[" + fmt.Sprint(cmd.Index) + "] " + cmd.Command + "</h3>")
		sb.WriteString("<p class='meta'>Timestamp: " + cmd.Timestamp.Format("2006-01-02 15:04:05.000") + "<br>")
		sb.WriteString("Full Command: <code>" + cmd.FullCommand + "</code></p>")
		sb.WriteString("<strong>Output Preview:</strong><pre>" + cmd.OutputPreview + "</pre>")
		if cmd.Signature != "" {
			sb.WriteString("<p class='meta'>Signature: <code>" + cmd.Signature + "</code></p>")
		}
		sb.WriteString("</div>")
	}

	sb.WriteString("<h2>Evidence Integrity</h2>")
	sb.WriteString("<table><tr><th>Path</th><th>SHA256</th><th>Captured</th></tr>")
	for _, f := range s.Evidence {
		sb.WriteString("<tr><td>" + f.OriginalPath + "</td><td><code>" + f.SHA256 + "</code></td><td>" + f.CapturedAt.Format("2006-01-02 15:04:05") + "</td></tr>")
	}
	sb.WriteString("</table><p>&nbsp;</p></body></html>")

	return sb.String(), nil
}
