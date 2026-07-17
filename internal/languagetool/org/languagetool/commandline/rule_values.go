package commandline

import (
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// parseRuleValues parses RULE_ID:value pairs (comma-separated blobs allowed).
func parseRuleValues(items []string) map[string]string {
	if len(items) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, item := range items {
		for _, part := range strings.Split(item, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			i := strings.IndexByte(part, ':')
			if i <= 0 || i == len(part)-1 {
				continue
			}
			id := strings.TrimSpace(part[:i])
			val := strings.TrimSpace(part[i+1:])
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
	lt := languagetool.NewJLanguageTool(lang)
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

// countErrorSeverityMatches counts matches whose SoftRuleMeta issue type maps to SARIF error.
func countErrorSeverityMatches(matches []*rules.RuleMatch) int {
	n := 0
	for _, m := range matches {
		if m == nil {
			continue
		}
		id := ruleIDOfMatch(m)
		_, _, issue, _ := languagetool.SoftRuleMeta(id)
		if languagetool.SeverityFromIssueType(issue) == "error" {
			n++
		}
	}
	return n
}
