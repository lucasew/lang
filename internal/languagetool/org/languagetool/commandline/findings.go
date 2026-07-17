package commandline

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Finding is the SPEC §2.2 finding contract for goldens and machine-readable output.
type Finding struct {
	Rule       string `json:"rule"`
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Location   string `json:"location"`
	Suggestion string `json:"suggestion,omitempty"`
	// URL is a soft community documentation link for the rule.
	URL string `json:"url,omitempty"`
	// Extra fields for golden precision (optional for consumers).
	Offset      int      `json:"offset,omitempty"`
	Length      int      `json:"length,omitempty"`
	File        string   `json:"file,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// MatchesToFindings converts rule matches to SPEC findings.
func MatchesToFindings(matches []*rules.RuleMatch, text, filename string) []Finding {
	if filename == "" || filename == "-" {
		filename = "stdin"
	}
	out := make([]Finding, 0, len(matches))
	for _, m := range matches {
		if m == nil {
			continue
		}
		id := ruleIDOfMatch(m)
		_, _, issue, _ := languagetool.SoftRuleMeta(id)
		if issue == "" {
			issue = "other"
		}
		sev := languagetool.SeverityFromIssueType(issue)
		line, col := LineColumnAt(text, m.FromPos)
		loc := fmt.Sprintf("%s:%d:%d", filename, line, col)
		sug := ""
		var all []string
		if reps := m.GetSuggestedReplacements(); len(reps) > 0 {
			sug = reps[0]
			all = append([]string(nil), reps...)
		}
		out = append(out, Finding{
			Rule:        id,
			Type:        issue,
			Severity:    sev,
			Message:     m.GetMessage(),
			Location:    loc,
			Suggestion:  sug,
			URL:         languagetool.SoftRuleURL(id, ""),
			Offset:      m.FromPos,
			Length:      m.ToPos - m.FromPos,
			File:        filename,
			Suggestions: all,
		})
	}
	return out
}

// WriteFindingsJSON writes a JSON array of findings (SPEC goldens shape).
func WriteFindingsJSON(w io.Writer, findings []Finding) error {
	if w == nil {
		return nil
	}
	if findings == nil {
		findings = []Finding{}
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(findings)
}

// CompareFindings returns human-readable diffs; empty string means equal (order-insensitive by rule+offset+message).
func CompareFindings(got, want []Finding) string {
	type key struct {
		Rule, Message string
		Offset, Length int
	}
	wantMap := map[key]Finding{}
	for _, f := range want {
		wantMap[key{f.Rule, f.Message, f.Offset, f.Length}] = f
	}
	gotMap := map[key]Finding{}
	for _, f := range got {
		gotMap[key{f.Rule, f.Message, f.Offset, f.Length}] = f
	}
	var b strings.Builder
	for k, wf := range wantMap {
		gf, ok := gotMap[k]
		if !ok {
			fmt.Fprintf(&b, "- missing: %s @%d %q\n", wf.Rule, wf.Offset, wf.Message)
			continue
		}
		if gf.Type != wf.Type {
			fmt.Fprintf(&b, "- %s @%d type: got %s want %s\n", wf.Rule, wf.Offset, gf.Type, wf.Type)
		}
		if gf.Severity != wf.Severity {
			fmt.Fprintf(&b, "- %s @%d severity: got %s want %s\n", wf.Rule, wf.Offset, gf.Severity, wf.Severity)
		}
	}
	for k, gf := range gotMap {
		if _, ok := wantMap[k]; !ok {
			fmt.Fprintf(&b, "- unexpected: %s @%d %q\n", gf.Rule, gf.Offset, gf.Message)
		}
	}
	return b.String()
}
