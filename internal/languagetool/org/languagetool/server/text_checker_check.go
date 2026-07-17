package server

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Check runs core rules for language on text and returns RemoteRuleMatch results.
// Uses Pipeline so disabled rules from query params can be applied.
func (t *TextChecker) Check(text, lang string, disabled []string) []RemoteRuleMatch {
	if lang == "" {
		lang = "en"
	}
	p := NewPipeline(NewPipelineSettings(lang, "check"))
	for _, id := range disabled {
		id = strings.TrimSpace(id)
		if id != "" {
			_ = p.DisableRuleID(id)
		}
	}
	_ = p.SetCleanOverlappingMatches(true)
	p.SetupFinished()
	locals := p.Check(text)
	ctxSize := DefaultContextSize
	if t != nil && t.ContextSize > 0 {
		ctxSize = t.ContextSize
	}
	return LocalMatchesToRemote(text, locals, ctxSize)
}

// CheckAndBuildJSON is a V2 convenience: Check + BuildResponse.
func (v *V2TextChecker) CheckAndBuildJSON(text, langCode, langName string, disabled []string) (string, error) {
	if v == nil {
		return "", nil
	}
	if langName == "" {
		langName = langCode
	}
	matches := v.Check(text, langCode, disabled)
	return v.BuildResponse(text, langCode, langName, matches)
}

// LocalMatchesToRemote maps cycle-free LocalMatch to API RemoteRuleMatch with context.
func LocalMatchesToRemote(text string, matches []languagetool.LocalMatch, contextSize int) []RemoteRuleMatch {
	if len(matches) == 0 {
		return nil
	}
	if contextSize <= 0 {
		contextSize = DefaultContextSize
	}
	out := make([]RemoteRuleMatch, 0, len(matches))
	for _, m := range matches {
		from, to := m.FromPos, m.ToPos
		if from < 0 {
			from = 0
		}
		if to < from {
			to = from
		}
		if to > len(text) {
			to = len(text)
		}
		start := from - contextSize
		if start < 0 {
			start = 0
		}
		end := to + contextSize
		if end > len(text) {
			end = len(text)
		}
		ctx := text[start:end]
		if ctx == "" {
			ctx = " " // NewRemoteRuleMatch requires non-empty context
		}
		msg := m.Message
		if msg == "" {
			msg = m.RuleID
		}
		if msg == "" {
			msg = "match"
		}
		ruleID := m.RuleID
		if ruleID == "" {
			ruleID = "UNKNOWN_RULE"
		}
		rm := NewRemoteRuleMatch(ruleID, msg, ctx, from-start, from, to-from)
		rm.Replacements = append([]string(nil), m.Suggestions...)
		out = append(out, *rm)
	}
	return out
}
