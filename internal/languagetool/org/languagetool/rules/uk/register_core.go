package uk

import (
	"os"
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
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

	// Java createDefaultSynthesizer → UkrainianSynthesizer for agreement suggestions.
	synth := discoverUkrainianSynthesizer()

	// Token agreement (Java order: VerbNoun, NounVerb, AdjNoun, PrepNoun, NumrNoun)
	vn := NewTokenAgreementVerbNounRule()
	vn.Synth = synth
	lt.AddRuleChecker(vn.GetID(), rules.AsSentenceCheckerSimple(vn.Match))

	nv := NewTokenAgreementNounVerbRule()
	lt.AddRuleChecker(nv.GetID(), rules.AsSentenceCheckerSimple(nv.Match))

	an := NewTokenAgreementAdjNounRule()
	an.Synth = synth
	lt.AddRuleChecker(an.GetID(), rules.AsSentenceCheckerSimple(an.Match))

	pn := NewTokenAgreementPrepNounRule()
	pn.Synth = synth
	lt.AddRuleChecker(pn.GetID(), rules.AsSentenceCheckerSimple(pn.Match))

	nn := NewTokenAgreementNumrNounRule()
	nn.Synth = synth
	lt.AddRuleChecker(nn.GetID(), rules.AsSentenceCheckerSimple(nn.Match))

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

// discoverUkrainianSynthesizer finds ukrainian_synth.dict (Java RESOURCE_FILENAME).
// Order: LANG_UK_SYNTH_DICT env, sibling of POS/speller dict dir.
func discoverUkrainianSynthesizer() synthesis.Synthesizer {
	if p := os.Getenv("LANG_UK_SYNTH_DICT"); p != "" {
		if s := synthesis.OpenBaseSynthesizerFromDictPath("uk", p); s != nil {
			return s
		}
	}
	// Sibling of official morfologik POS / speller resource dir
	for _, dictName := range []string{UkrainianSpellerDict, "ukrainian.dict"} {
		if pos := morfologik.DiscoverLanguageDict(dictName); pos != "" {
			cand := filepath.Join(filepath.Dir(pos), "ukrainian_synth.dict")
			if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
				if s := synthesis.OpenBaseSynthesizerFromDictPath("uk", cand); s != nil {
					return s
				}
			}
		}
	}
	return nil
}
