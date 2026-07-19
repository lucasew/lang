package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreUkrainianRules installs Java Ukrainian.getRelevantRules ports
// (layout, token agreement, replace, speller). Pattern grammar.xml still
// loaded separately via GetRuleFileNames when LANG_USE_UPSTREAM_GRAMMAR=1.
func RegisterCoreUkrainianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Java Ukrainian.IGNORED_CHARS: soft hyphen + combining acute.
	lt.IgnoredCharacters = languagetool.UkrainianIgnoredCharactersRegex

	// Layout: Java skips DoublePunctuationRule; UK comma exceptions for en/em dash;
	// Ukrainian uppercase list-item а) б) exceptions.
	ukComma := NewUkrainianCommaWhitespaceRule(nil)
	ukUpper := NewUkrainianUppercaseSentenceStartRule(nil)
	rules.RegisterSharedLayoutRulesWithOptions(lt, "uk", rules.SharedLayoutOptions{
		// Java getRelevantRules does not include DoublePunctuationRule (TODO in Java).
		SkipDoublePunctuation: true,
		CommaException:        ukComma.IsException,
		UppercaseMatchList: func(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
			return ukUpper.MatchList(sentences)
		},
	})

	// word-repeat (+ beginning as text-level; Java only has WordRepeatRule in list,
	// but WordRepeatBeginning is common in Go language packs and already here).
	wr := NewUkrainianWordRepeatRule(map[string]string{"repetition": "Повтор слова"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Три речення поспіль починаються одним словом.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// medium/high priority Java rules
	typo := NewTypographyRule(nil)
	lt.AddRuleChecker(typo.GetID(), rules.AsSentenceCheckerSimple(typo.Match))

	hidden := NewHiddenCharacterRule(nil)
	lt.AddRuleChecker(hidden.GetID(), rules.AsSentenceCheckerSimple(hidden.Match))

	// Java createDefaultSpellingRule → MorfologikUkrainianSpellerRule.
	sp := NewMorfologikUkrainianSpellerRule()
	if p := morfologik.DiscoverLanguageDict(UkrainianSpellerDict); p != "" {
		if WireUkrainianFilterSpeller(p) {
			inner := FilterDictIsMisspelledUK
			sp.IsMisspelled = func(word string) bool {
				return sp.ukIsMisspelled(word, inner)
			}
		}
	}
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	// high priority
	hyphen := NewMissingHyphenRule(nil)
	lt.AddRuleChecker(hyphen.GetID(), rules.AsSentenceCheckerSimple(hyphen.Match))

	// Token agreement (Java order: VerbNoun, NounVerb, AdjNoun, PrepNoun, NumrNoun)
	for _, reg := range []struct {
		id string
		fn func(*languagetool.AnalyzedSentence) []*rules.RuleMatch
	}{
		{TokenAgreementVerbNounRuleID, NewTokenAgreementVerbNounRule().Match},
		{TokenAgreementNounVerbRuleID, NewTokenAgreementNounVerbRule().Match},
		{TokenAgreementAdjNounRuleID, NewTokenAgreementAdjNounRule().Match},
		{TokenAgreementPrepNounRuleID, NewTokenAgreementPrepNounRule().Match},
		{TokenAgreementNumrNounRuleID, NewTokenAgreementNumrNounRule().Match},
	} {
		id, fn := reg.id, reg.fn
		lt.AddRuleChecker(id, rules.AsSentenceCheckerSimple(fn))
	}

	mixed := NewMixedAlphabetsRule(nil)
	lt.AddRuleChecker(mixed.GetID(), rules.AsSentenceCheckerSimple(mixed.Match))

	// Official replace tables (Java order: Soft, Renamed, SimpleReplace).
	ss := NewSimpleReplaceSoftRule(nil)
	lt.AddRuleChecker(ss.GetID(), rules.AsSentenceCheckerSimple(ss.Match))
	rn := NewSimpleReplaceRenamedRule(nil)
	lt.AddRuleChecker(rn.GetID(), rules.AsSentenceCheckerSimple(rn.Match))
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
}
