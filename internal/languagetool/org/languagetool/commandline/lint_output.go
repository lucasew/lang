package commandline

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// WriteLintMatches writes SPEC §2.2 text columns:
// location, severity, type, rule, message, suggestion.
func WriteLintMatches(w io.Writer, matches []*rules.RuleMatch, text, filename string) error {
	if w == nil {
		return nil
	}
	if filename == "" || filename == "-" {
		filename = "stdin"
	}
	tw := tabwriter.NewWriter(w, 0, 4, 1, ' ', 0)
	_, _ = fmt.Fprintln(tw, "location\tseverity\ttype\trule\tmessage\tsuggestion")
	for _, m := range matches {
		if m == nil {
			continue
		}
		id := ruleIDOfMatch(m)
		_, _, issue, _ := languagetool.SoftRuleMeta(id)
		sev := languagetool.SeverityFromIssueType(issue)
		if issue == "" {
			issue = "other"
		}
		line, col := LineColumnAt(text, m.FromPos)
		loc := fmt.Sprintf("%s:%d:%d", filename, line, col)
		sug := ""
		if reps := m.GetSuggestedReplacements(); len(reps) > 0 {
			sug = reps[0]
		}
		msg := strings.ReplaceAll(m.GetMessage(), "\t", " ")
		msg = strings.ReplaceAll(msg, "\n", " ")
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", loc, sev, issue, id, msg, sug)
	}
	return tw.Flush()
}
