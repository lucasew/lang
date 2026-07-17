package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// ToLocalMatches converts RuleMatch list to cycle-free LocalMatch for JLanguageTool.Check.
func ToLocalMatches(ms []*RuleMatch) []languagetool.LocalMatch {
	if len(ms) == 0 {
		return nil
	}
	out := make([]languagetool.LocalMatch, 0, len(ms))
	for _, m := range ms {
		if m == nil {
			continue
		}
		lm := languagetool.LocalMatch{
			FromPos:     m.GetFromPos(),
			ToPos:       m.GetToPos(),
			Message:     m.GetMessage(),
			Suggestions: append([]string(nil), m.GetSuggestedReplacements()...),
		}
		if g, ok := m.Rule.(interface{ GetID() string }); ok {
			lm.RuleID = g.GetID()
		}
		out = append(out, lm)
	}
	return out
}

// AsSentenceChecker wraps a Match(sentence)([]*RuleMatch, error) as SentenceChecker.
func AsSentenceChecker(match func(*languagetool.AnalyzedSentence) ([]*RuleMatch, error)) languagetool.SentenceChecker {
	return func(s *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		if match == nil {
			return nil
		}
		ms, err := match(s)
		if err != nil {
			return nil
		}
		return ToLocalMatches(ms)
	}
}

// AsSentenceCheckerSimple wraps Match(sentence) []*RuleMatch (no error).
func AsSentenceCheckerSimple(match func(*languagetool.AnalyzedSentence) []*RuleMatch) languagetool.SentenceChecker {
	return func(s *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		if match == nil {
			return nil
		}
		return ToLocalMatches(match(s))
	}
}

// RegisterCoreEnglishRules installs whitespace, double punctuation, uppercase start,
// and word-repeat inject onto lt for integration smokes.
func RegisterCoreEnglishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Text-level rules accept sentence slices; wrap single sentence for Check path.
	ws := NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	dp := NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), AsSentenceCheckerSimple(dp.Match))

	up := NewUppercaseSentenceStartRule(nil, "en")
	lt.AddRuleChecker(up.GetID(), AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	// soft injects for gaps (no full XML rules)
	lt.AddRuleChecker("WORD_REPEAT_RULE", languagetool.SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.AddRuleChecker("EN_A_VS_AN", languagetool.SimpleAvsAnChecker())
	lt.AddRuleChecker("UNPAIRED_BRACKETS", languagetool.SimpleUnpairedBracketsChecker())
}

// RegisterPatternRule wires a PatternRule into Check (simplified matcher).
func RegisterPatternRule(lt *languagetool.JLanguageTool, match func(*languagetool.AnalyzedSentence) ([]*RuleMatch, error), id string) {
	if lt == nil || match == nil {
		return
	}
	if id == "" {
		id = "PATTERN_RULE"
	}
	lt.AddRuleChecker(id, AsSentenceChecker(match))
}
