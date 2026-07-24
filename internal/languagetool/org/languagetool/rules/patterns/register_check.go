package patterns

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterPatternRule wires a PatternRule into JLanguageTool.Check.
// XML default="off" → MarkDefaultOff; default="temp_off" → MarkDefaultTempOff
// (Java setDefaultOff / setDefaultTempOff; re-enable with EnableRule / EnableTempOffRules).
func RegisterPatternRule(lt *languagetool.JLanguageTool, pr *PatternRule) {
	if lt == nil || pr == nil {
		return
	}
	id := pr.GetID()
	if id == "" {
		id = "PATTERN_RULE"
	}
	if pr.DefaultTempOff {
		lt.MarkDefaultTempOff(id)
	} else if pr.DefaultOff {
		lt.MarkDefaultOff(id)
	}
	lt.AddRuleChecker(id, rules.AsSentenceChecker(pr.Match))
}

// RegisterLoadedPatternRules registers all PatternRules from a PatternRuleHandler.
func RegisterLoadedPatternRules(lt *languagetool.JLanguageTool, h *PatternRuleHandler) {
	if lt == nil || h == nil {
		return
	}
	// Java Category.isDefaultOff from XML category default="off".
	for id, cat := range h.Categories {
		if cat != nil && cat.IsDefaultOff() {
			lt.MarkCategoryDefaultOff(id)
		}
	}
	for _, pr := range h.LoadedPatternRules {
		RegisterPatternRule(lt, pr)
	}
}

// RegisterTokenSequence registers a simple surface-token sequence pattern rule.
// suggestion, if non-empty, is attached as the sole suggested replacement on matches.
func RegisterTokenSequence(lt *languagetool.JLanguageTool, id, lang string, tokens []string, message, suggestion string) {
	if lt == nil || len(tokens) == 0 {
		return
	}
	if id == "" {
		id = "PATTERN_SEQUENCE"
	}
	pts := make([]*PatternToken, len(tokens))
	for i, t := range tokens {
		pts[i] = Token(t)
	}
	if message == "" {
		message = "Possible pattern match"
	}
	pr := NewPatternRule(id, lang, pts, message, message, "")
	lt.AddRuleChecker(id, func(s *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		ms, err := pr.Match(s)
		if err != nil || len(ms) == 0 {
			return nil
		}
		if suggestion != "" {
			for _, m := range ms {
				if m != nil && len(m.GetSuggestedReplacements()) == 0 {
					m.SetSuggestedReplacement(suggestion)
				}
			}
		}
		out := rules.ToLocalMatches(ms)
		// Token-sequence injects are grammar patterns (Java Categories.GRAMMAR / ITS grammar)
		// when the rule does not carry its own category metadata.
		for i := range out {
			if out[i].IssueType == "" {
				out[i].IssueType = "grammar"
			}
			if out[i].CategoryID == "" {
				out[i].CategoryID = "GRAMMAR"
				out[i].CategoryName = "Grammar"
			}
		}
		return out
	})
}

// RegisterTokenSequences registers multiple token-sequence patterns.
func RegisterTokenSequences(lt *languagetool.JLanguageTool, lang string, specs []TokenSequenceSpec) {
	for _, s := range specs {
		RegisterTokenSequence(lt, s.ID, lang, s.Tokens, s.Message, s.Suggestion)
	}
}

// TokenSequenceSpec is a compact pattern definition for Check injects.
type TokenSequenceSpec struct {
	ID         string
	Tokens     []string
	Message    string
	Suggestion string
}
