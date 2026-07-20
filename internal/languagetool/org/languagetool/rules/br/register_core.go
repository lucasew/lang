package br

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreBretonRules ports Breton.getRelevantRules / createDefaultSpellingRule.
// Java: CommaWhitespace, DoublePunctuation, MorfologikBretonSpeller, UppercaseSentenceStart,
// MultipleWhitespace, SentenceWhitespace, TopoReplace — no WordRepeat / WordRepeatBeginning /
// BretonCompoundRule (those types exist but are not in getRelevantRules).
func RegisterCoreBretonRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "br")

	// Official topo.txt place-name replace (embedded from upstream).
	tr := NewTopoReplaceRule(nil)
	lt.AddRuleChecker(tr.GetID(), rules.AsSentenceCheckerSimple(tr.Match))

	// Java createDefaultSpellingRule → MorfologikBretonSpellerRule.
	// Always full Match (IgnoreTaggedWords + hyphen tokenizingPattern).
	sp := NewMorfologikBretonSpellerRule()
	if p := morfologik.DiscoverLanguageDict(MorfologikBretonSpellerRuleDict); p != "" {
		// Binary CFSA2 optional — fail-closed map Words when missing.
		_ = p
	}
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
