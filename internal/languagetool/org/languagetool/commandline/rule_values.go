package commandline

import (
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// parseRuleValues ports TextChecker.getRuleValues / CLI --ruleValues:
// split on "," and ":" without invent per-field TrimSpace (server twin).
func parseRuleValues(items []string) map[string]string {
	if len(items) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, item := range items {
		if item == "" {
			continue
		}
		for _, part := range strings.Split(item, ",") {
			if part == "" {
				continue
			}
			i := strings.IndexByte(part, ':')
			if i < 0 || i == len(part)-1 {
				continue
			}
			id := part[:i]
			val := part[i+1:]
			if id != "" && val != "" {
				out[strings.ToUpper(id)] = val
			}
		}
	}
	return out
}

// applyCLIRuleValues re-runs long-sentence detection with a custom max when configured.
func applyCLIRuleValues(lang, text string, existing []*rules.RuleMatch, raw []string) []*rules.RuleMatch {
	vals := parseRuleValues(raw)
	if len(vals) == 0 {
		return existing
	}
	maxStr, ok := vals["TOO_LONG_SENTENCE"]
	if !ok {
		maxStr, ok = vals["LONG_SENTENCE_RULE"]
	}
	if !ok {
		return existing
	}
	maxWords, err := strconv.Atoi(maxStr)
	if err != nil || maxWords <= 0 {
		return existing
	}
	out := make([]*rules.RuleMatch, 0, len(existing))
	for _, m := range existing {
		if m == nil {
			continue
		}
		id := strings.ToUpper(ruleIDOfMatch(m))
		if strings.Contains(id, "LONG_SENTENCE") || strings.Contains(id, "TOO_LONG_SENTENCE") {
			continue
		}
		out = append(out, m)
	}
	ls := rules.NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "This sentence is too long (%d words)",
	}, maxWords)
	// LongSentenceRule is Tag.picky; explicit ruleValues threshold uses Level.PICKY.
	lt := languagetool.NewJLanguageTool(lang)
	lt.Level = languagetool.LevelPicky
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		if id != ls.GetID() {
			lt.DisableRule(id)
		}
	}
	sent := languagetool.AnalyzePlain(text)
	extra := rules.FromLocalMatches(lt.Check(text), sent)
	return append(out, extra...)
}

// severityRank maps SARIF levels for --fail-on comparisons (higher = worse).
func severityRank(sev string) int {
	// CLI flag surface: String.trim-like edge only (ASCII ≤U+0020).
	switch strings.ToLower(tools.JavaStringTrim(sev)) {
	case "error":
		return 3
	case "warning":
		return 2
	case "note":
		return 1
	default:
		return 0
	}
}

// countFailOnMatches counts matches at or above the fail-on severity threshold.
// failOn is error|warning|note (default error).
func countFailOnMatches(matches []*rules.RuleMatch, failOn string) int {
	threshold := severityRank(failOn)
	if threshold == 0 {
		threshold = severityRank("error")
	}
	n := 0
	for _, m := range matches {
		if m == nil {
			continue
		}
		// Prefer issue type from the match (set by the rule); RuleMeta is fallback only.
		issue := m.IssueType
		if issue == "" {
			id := ruleIDOfMatch(m)
			_, _, issue, _ = languagetool.RuleMeta(id)
		}
		sev := languagetool.SeverityFromIssueType(issue)
		if severityRank(sev) >= threshold {
			n++
		}
	}
	return n
}

// countErrorSeverityMatches counts matches whose issue type maps to SARIF error.
func countErrorSeverityMatches(matches []*rules.RuleMatch) int {
	return countFailOnMatches(matches, "error")
}
