package uk

import (
	"os"
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// RegisterCoreUkrainianRules ports Ukrainian.getRelevantRules.
// Java list only — no invent SharedLayout extras (no double-punct / unpaired /
// sentence-whitespace / empty-line / long-paragraph / paragraph-repeat / etc.).
func RegisterCoreUkrainianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Java Ukrainian.IGNORED_CHARS: soft hyphen + combining acute.
	lt.IgnoredCharacters = languagetool.UkrainianIgnoredCharactersRegex

	// lower priority
	ukComma := NewUkrainianCommaWhitespaceRule(nil)
	lt.AddRuleChecker(ukComma.GetID(), rules.AsSentenceCheckerSimple(ukComma.Match))

	ukUpper := NewUkrainianUppercaseSentenceStartRule(nil)
	lt.AddRuleChecker(ukUpper.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ukUpper.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	// Java UkrainianWordRepeatRule only — no WordRepeatBeginning.
	wr := NewUkrainianWordRepeatRule(map[string]string{"repetition": "Повтор слова"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	typo := NewTypographyRule(nil)
	lt.AddRuleChecker(typo.GetID(), rules.AsSentenceCheckerSimple(typo.Match))

	hidden := NewHiddenCharacterRule(nil)
	lt.AddRuleChecker(hidden.GetID(), rules.AsSentenceCheckerSimple(hidden.Match))

	// medium priority — Java createDefaultSpellingRule → MorfologikUkrainianSpellerRule.
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

	synth := discoverUkrainianSynthesizer()

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

	ss := NewSimpleReplaceSoftRule(nil)
	lt.AddRuleChecker(ss.GetID(), rules.AsSentenceCheckerSimple(ss.Match))
	rn := NewSimpleReplaceRenamedRule(nil)
	lt.AddRuleChecker(rn.GetID(), rules.AsSentenceCheckerSimple(rn.Match))
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
}

// discoverUkrainianSynthesizer finds ukrainian_synth.dict (Java RESOURCE_FILENAME).
func discoverUkrainianSynthesizer() synthesis.Synthesizer {
	if p := os.Getenv("LANG_UK_SYNTH_DICT"); p != "" {
		if s := synthesis.OpenBaseSynthesizerFromDictPath("uk", p); s != nil {
			return s
		}
	}
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
