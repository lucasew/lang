package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// ToLocalMatches converts RuleMatch list to cycle-free LocalMatch for JLanguageTool.Check.
// Copies rule ID / category / ITS / description when the rule implements the Java getters.
func ToLocalMatches(ms []*RuleMatch) []languagetool.LocalMatch {
	if len(ms) == 0 {
		return nil
	}
	out := make([]languagetool.LocalMatch, 0, len(ms))
	for _, m := range ms {
		if m == nil {
			continue
		}
		// Java SwissGerman.filterRuleMatches uses sentence.substring(from,to);
		// setOriginalErrorStr fills the same surface when not already set.
		if m.GetOriginalErrorStr() == "" {
			m.SetOriginalErrorStr()
		}
		lm := languagetool.LocalMatch{
			FromPos:      m.GetFromPos(),
			ToPos:        m.GetToPos(),
			Message:      m.GetMessage(),
			ShortMessage: m.GetShortMessage(),
			Suggestions:  append([]string(nil), m.GetSuggestedReplacements()...),
			IssueType:    m.IssueType,
			CategoryID:   m.CategoryID,
			CategoryName: m.CategoryName,
			// Java RuleMatch.getUrl: match URL first, else rule URL.
			URL: m.GetURL(),
			// Surface under marker (Swiss AI ss→ß skip, AdaptSuggestionsFilter, …).
			OriginalErrorStr: m.GetOriginalErrorStr(),
			// Sentence-relative span for CleanOverlapping isPunctuationOnlyChange.
			FromPosSentence: m.GetFromPosSentence(),
			ToPosSentence:   m.GetToPosSentence(),
		}
		if m.Sentence != nil {
			lm.SentenceText = m.Sentence.GetText()
		}
		// Java JSON/API use RuleMatch.getSpecificRuleId() as the public rule id
		// (specificRuleId when set, else rule.getId()) — e.g. DE_REPEATEDWORDS_AUSSERDEM.
		lm.RuleID = m.GetSpecificRuleId()
		if g, ok := m.Rule.(interface{ GetDescription() string }); ok {
			lm.Description = g.GetDescription()
		}
		if g, ok := m.Rule.(interface{ GetCategory() *Category }); ok {
			if cat := g.GetCategory(); cat != nil {
				if lm.CategoryID == "" {
					lm.CategoryID = cat.GetID().String()
				}
				if lm.CategoryName == "" {
					lm.CategoryName = cat.GetName()
				}
			}
		}
		if g, ok := m.Rule.(interface{ GetLocQualityIssueType() ITSIssueType }); ok {
			if it := g.GetLocQualityIssueType(); it != "" && lm.IssueType == "" {
				lm.IssueType = string(it)
			}
		}
		// Java RuleMatch.getUrl falls back to rule.getUrl() when match URL empty.
		if lm.URL == "" {
			if g, ok := m.Rule.(interface{ GetURL() string }); ok {
				lm.URL = g.GetURL()
			}
		}
		// Rule.getTags → LocalMatch.Tags (JSON rule.tags) + IsPicky for demotion/merge.
		if g, ok := m.Rule.(interface{ GetTags() []Tag }); ok {
			if tags := g.GetTags(); len(tags) > 0 {
				lm.Tags = make([]string, len(tags))
				for i, t := range tags {
					lm.Tags[i] = string(t)
					if t == TagPicky {
						lm.IsPicky = true
					}
				}
			}
		} else if g, ok := m.Rule.(interface{ HasTag(Tag) bool }); ok {
			lm.IsPicky = g.HasTag(TagPicky)
			if lm.IsPicky {
				lm.Tags = []string{string(TagPicky)}
			}
		}
		// Java Rule.isDefaultTempOff → JSON rule.tempOff.
		if g, ok := m.Rule.(interface{ IsDefaultTempOff() bool }); ok {
			lm.TempOff = g.IsDefaultTempOff()
		}
		// Java Rule.isIncludedInErrorsCorrectedAllAtOnce for punctuation-only overlap preference.
		if g, ok := m.Rule.(interface{ IsIncludedInErrorsCorrectedAllAtOnce() bool }); ok {
			lm.IncludedInErrorsCorrectedAllAtOnce = g.IsIncludedInErrorsCorrectedAllAtOnce()
		}
		// Premium flag: rule method, else DefaultPremium (Java Premium.get().isPremiumRule),
		// else RuleID contains "PREMIUM" for LocalMatch-only paths without a Premium registry.
		if g, ok := m.Rule.(interface{ IsPremium() bool }); ok {
			lm.IsPremium = g.IsPremium()
		} else if g, ok := m.Rule.(interface{ GetPremium() bool }); ok {
			lm.IsPremium = g.GetPremium()
		}
		if !lm.IsPremium && lm.RuleID != "" {
			if languagetool.DefaultPremium != nil && languagetool.DefaultPremium.IsPremiumRule(lm.RuleID) {
				lm.IsPremium = true
			} else if strings.Contains(lm.RuleID, "PREMIUM") {
				lm.IsPremium = true
			}
		}
		// Java Rule.estimateContextForSureMatch (default 0; TextLevelRuleBase → -1).
		if g, ok := m.Rule.(interface{ EstimateContextForSureMatch() int }); ok {
			lm.EstimateContextForSureMatch = g.EstimateContextForSureMatch()
		}
		// Java Rule.getPriority (XML prio=); applyRulePriorities may still overlay by id.
		if g, ok := m.Rule.(interface{ GetPriority() int }); ok {
			if p := g.GetPriority(); p != 0 {
				lm.Priority = p
			}
		}
		// Tone tags + goalSpecific for isRuleActiveForLevelAndToneTags.
		if g, ok := m.Rule.(interface {
			GetToneTags() []languagetool.ToneTag
		}); ok {
			if tags := g.GetToneTags(); len(tags) > 0 {
				lm.ToneTags = append([]languagetool.ToneTag(nil), tags...)
			}
		}
		if g, ok := m.Rule.(interface{ IsGoalSpecific() bool }); ok {
			lm.GoalSpecific = g.IsGoalSpecific()
		}
		// RuleMeta fallback when rule getters left category/ITS empty (CLI/API parity).
		// Only apply known Java families — skip uncategorized invent for unknown IDs.
		if lm.RuleID != "" && (lm.CategoryID == "" || lm.IssueType == "" || lm.Description == "" || lm.ShortMessage == "") {
			catID, catName, issue, short := languagetool.RuleMeta(lm.RuleID)
			if issue != "" && issue != "uncategorized" {
				if lm.CategoryID == "" {
					lm.CategoryID = catID
				}
				if lm.CategoryName == "" {
					lm.CategoryName = catName
				}
				if lm.IssueType == "" {
					lm.IssueType = issue
				}
				if lm.ShortMessage == "" && short != "" {
					lm.ShortMessage = short
				}
				if lm.Description == "" {
					if d := languagetool.RuleDescription(lm.RuleID); d != "" && d != lm.RuleID {
						lm.Description = d
					}
				}
			}
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

// AsTextLevelChecker wraps MatchList(sentences) []*RuleMatch as TextLevelChecker.
// Offsets are already document-relative (rules accumulate GetCorrectedTextLength).
// Java TextLevelRule.estimateContextForSureMatch is always -1 — apply when the rule
// object did not already set a value via EstimateContextForSureMatch().
func AsTextLevelChecker(matchList func([]*languagetool.AnalyzedSentence) []*RuleMatch) languagetool.TextLevelChecker {
	return func(sents []*languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		if matchList == nil {
			return nil
		}
		ms := ToLocalMatches(matchList(sents))
		for i := range ms {
			// Default Java TextLevelRule → -1 when rule left zero (sentence-rule default).
			// Rules that implement EstimateContextForSureMatch already filled the field.
			if ms[i].EstimateContextForSureMatch == 0 {
				// Distinguish unset 0 from explicit 0: text-level Java always returns -1.
				ms[i].EstimateContextForSureMatch = -1
			}
		}
		return ms
	}
}

// RegisterCoreEnglishRules is a legacy thin entry for rules-package tests.
// Prefer rules/en.RegisterCoreEnglishLanguageRules (Java English.getRelevantRules).
// Does not invent SharedLayout extras (WHITESPACE_PUNCTUATION etc.).
func RegisterCoreEnglishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Minimal EN layout subset used by rules-package adapter tests only.
	// Full Java list is rules/en.RegisterCoreEnglishLanguageRules.
	cw := NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), AsSentenceCheckerSimple(cw.Match))
	dp := NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), AsSentenceCheckerSimple(dp.Match))
	up := NewUppercaseSentenceStartRule(nil, "en")
	lt.AddRuleChecker(up.GetID(), AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))
	ws := NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))
	sw := NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(sw.GetID(), AsTextLevelChecker(sw.MatchList))
	el := NewEmptyLineRule(map[string]string{"empty_line_rule_msg": "Empty line"})
	lt.AddTextLevelRuleChecker(el.GetID(), AsTextLevelChecker(el.MatchList))
	if el.IsDefaultOff() {
		lt.MarkDefaultOff(el.GetID())
	}
	// Java EnglishWordRepeatRule is in package en (ENGLISH_WORD_REPEAT_RULE).
	// Do not invent default WORD_REPEAT_RULE here.
	if languagetool.PreferredAvsAnChecker != nil {
		lt.AddRuleChecker("EN_A_VS_AN", languagetool.PreferredAvsAnChecker)
	}
}

// RegisterCoreRules is a legacy dispatcher for unknown short codes.
// No invent SharedLayout for unlisted languages — Java only registers rules
// listed in that language's getRelevantRules. Known languages use corepack.
func RegisterCoreRules(lt *languagetool.JLanguageTool, langCode string) {
	if lt == nil {
		return
	}
	base := langCode
	if i := indexByteLocal(langCode, '-'); i > 0 {
		base = langCode[:i]
	}
	switch base {
	case "en":
		RegisterCoreEnglishRules(lt)
	default:
		// No invent layout for unknown codes (faithful-port: leaves → dedicated packs).
		_ = base
	}
}

func indexByteLocal(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
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
