package br

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreBretonRules ports Breton.getRelevantRules / createDefaultSpellingRule.
// Java list only (no invent SharedLayout extras: unpaired brackets, empty line, long
// paragraph, paragraph-repeat beginning, whitespace-before-punct, etc.).
func RegisterCoreBretonRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}

	// Java order: CommaWhitespace, DoublePunctuation, Morfologik, Uppercase,
	// MultipleWhitespace, SentenceWhitespace, TopoReplace.
	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	// Java createDefaultSpellingRule → MorfologikBretonSpellerRule.
	sp := NewMorfologikBretonSpellerRule()
	if p := morfologik.DiscoverLanguageDict(MorfologikBretonSpellerRuleDict); p != "" {
		_ = p // binary CFSA2 optional — fail-closed map Words when missing
	}
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	up := rules.NewUppercaseSentenceStartRule(nil, "br")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	sw := rules.NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(sw.GetID(), rules.AsTextLevelChecker(sw.MatchList))

	// Official topo.txt place-name replace (embedded from upstream).
	tr := NewTopoReplaceRule(nil)
	lt.AddRuleChecker(tr.GetID(), rules.AsSentenceCheckerSimple(tr.Match))
}
