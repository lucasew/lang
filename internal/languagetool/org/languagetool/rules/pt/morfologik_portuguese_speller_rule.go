package pt

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	ptsynthesis "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	taggingpt "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pt"
	pttok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Portuguese speller variants — IDs and dict paths from
// org.languagetool.rules.pt.MorfologikPortugueseSpellerRule.
const (
	// getId(): "MORFOLOGIK_RULE_" + shortCodeWithCountryAndVariant uppercased.
	MorfologikPortuguesePTSpellerRuleID = "MORFOLOGIK_RULE_PT_PT"
	MorfologikPortugueseBRSpellerRuleID = "MORFOLOGIK_RULE_PT_BR"
	MorfologikPortugueseAOSpellerRuleID = "MORFOLOGIK_RULE_PT_AO"
	MorfologikPortugueseMZSpellerRuleID = "MORFOLOGIK_RULE_PT_MZ"
	// Java: dictFilepath = "/pt/spelling/" + getDictFilename() + ".dict"
	// pt-PT → pt-PT-90; pt-BR → pt-BR; pt-AO/pt-MZ → pt-PT-45.
	PortuguesePTDict   = "/pt/spelling/pt-PT-90.dict"
	PortugueseBRDict   = "/pt/spelling/pt-BR.dict"
	PortuguesePT45Dict = "/pt/spelling/pt-PT-45.dict"
	// EnglishIgnorePOS ports POS tag skipped in Java getRuleMatches.
	englishIgnorePOS = "_english_ignore_"
)

// Java static PortugueseTagger / PortugueseSynthesizer.INSTANCE — loaded once when assets exist.
var (
	ptSpellerTaggerOnce sync.Once
	ptSpellerTagger     *taggingpt.PortugueseTagger

	ptSpellerSynthOnce sync.Once
	ptSpellerSynth     *ptsynthesis.PortugueseSynthesizer
)

func sharedPortugueseSpellerTagger() *taggingpt.PortugueseTagger {
	ptSpellerTaggerOnce.Do(func() {
		p := morfologik.DiscoverLanguageDict(taggingpt.PortugueseDictPath)
		if p == "" {
			p = spelling.DiscoverSpellingResource(strings.TrimPrefix(taggingpt.PortugueseDictPath, "/"))
		}
		if p == "" {
			return
		}
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		ptSpellerTagger = taggingpt.NewPortugueseTagger(mt)
		// Java PortugueseWordTokenizer uses PortugueseTagger for hyphen compounds.
		tg := ptSpellerTagger
		pttok.IsTaggedPT = func(s string) bool {
			if tg == nil || s == "" {
				return false
			}
			atrs := tg.Tag([]string{s})
			return len(atrs) > 0 && atrs[0] != nil && atrs[0].IsTagged()
		}
	})
	return ptSpellerTagger
}

func sharedPortugueseSpellerSynth() *ptsynthesis.PortugueseSynthesizer {
	ptSpellerSynthOnce.Do(func() {
		p := morfologik.DiscoverLanguageDict(ptsynthesis.PortugueseSynthDict)
		if p == "" {
			p = spelling.DiscoverSpellingResource(strings.TrimPrefix(ptsynthesis.PortugueseSynthDict, "/"))
		}
		if p == "" {
			return
		}
		ptSpellerSynth = ptsynthesis.OpenPortugueseSynthesizerFromDictPath(p)
	})
	return ptSpellerSynth
}

// MorfologikPortugueseSpellerRule ports rules.pt.MorfologikPortugueseSpellerRule.
// Word lists: pt/ignore.txt, pt/prohibit.txt, pt/spelling.txt, pt/multiwords.txt.
// getRuleMatches post-filters: _english_ignore_, titlecase-hyphen, diaeresis,
// European 1PL past (pt-BR), dialect surface/lemma map, compound-element suggestions,
// do_not_suggest filter. Clitic-verb needs TagPOS (wired when portuguese.dict present).
type MorfologikPortugueseSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	VariantCode string
	// dialectMap ports dialectAlternationMapping for this variant.
	dialectMap map[string]string
	// TagPOS optional PortugueseTagger.tag surface → POS tags (clitic verb + dialect lemma).
	TagPOS func(word string) []string
	// TagLemma optional lemmas for dialect/clitic dialect invalidation.
	TagLemma func(word string) []string
	// SynthesizeForPosTags ports PortugueseSynthesizer.synthesizeForPosTags (dialectAlternative).
	// Nil → lemma expansion arm fail-closed (surface map only).
	SynthesizeForPosTags func(lemma string, acceptTag func(string) bool) []string
}

// dictFilenameForVariant ports getDictFilename().
func dictFilenameForVariant(variantCode string) string {
	switch variantCode {
	case "pt-BR":
		return "pt-BR"
	case "pt-PT":
		return "pt-PT-90"
	case "pt-AO", "pt-MZ":
		return "pt-PT-45"
	default:
		return "pt-BR" // Java default
	}
}

// DictPathForVariant ports dictFilepath construction.
func DictPathForVariant(variantCode string) string {
	return "/pt/spelling/" + dictFilenameForVariant(variantCode) + ".dict"
}

// RuleIDForVariant ports getId().
func RuleIDForVariant(variantCode string) string {
	// Java: language.getShortCodeWithCountryAndVariant().replace("-", "_").toUpperCase()
	code := variantCode
	if code == "" {
		code = "pt"
	}
	return "MORFOLOGIK_RULE_" + strings.ToUpper(strings.ReplaceAll(code, "-", "_"))
}

func NewMorfologikPortugueseSpellerRule(variantCode, dict, id string) *MorfologikPortugueseSpellerRule {
	if dict == "" {
		dict = DictPathForVariant(variantCode)
	}
	if id == "" {
		id = RuleIDForVariant(variantCode)
	}
	// Java: if language short is bare "pt", use defaultLanguageVariant for spellerLanguage.
	spellerVariant := variantCode
	if spellerVariant == "pt" {
		spellerVariant = "pt-PT" // Java Portuguese default variant for bare pt
	}
	r := &MorfologikPortugueseSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(id, "pt", dict, nil),
		VariantCode:           spellerVariant,
		dialectMap:            loadDialectAlternationMapping(spellerVariant),
	}
	// Java path overrides: pt/ignore.txt, pt/prohibit.txt; additional = global + pt/spelling + multiwords.
	if r.SpellingCheckRule != nil {
		r.GetIgnoreFileNameFn = func() string { return "pt/ignore.txt" }
		r.GetProhibitFileNameFn = func() string { return "pt/prohibit.txt" }
		// getSpellingFileName stays default (may miss); spelling is in additional list.
		r.GetAdditionalSpellingFileNamesFn = func() []string {
			return []string{spelling.GlobalSpellingFile, "pt/spelling.txt", "pt/multiwords.txt"}
		}
		spelling.ReapplyDefaultSpellingWordLists(r.SpellingCheckRule)
		// Java filterNoSuggestWords(doNotSuggestWords)
		r.FilterNoSuggestWordsFn = filterDoNotSuggest
		// Java getAdditionalTopSuggestions → abbreviation + "."
		r.GetAdditionalTopSuggestionsFn = func(existing []string, word string) []string {
			if isAbbreviation(word) {
				return []string{word + "."}
			}
			return nil
		}
	}
	// Java getRuleMatches: if tokens[idx].hasPosTag("_english_ignore_") return empty.
	r.SkipTokenFn = func(tok *languagetool.AnalyzedTokenReadings) bool {
		return tok != nil && tok.HasPosTag(englishIgnorePOS)
	}
	// Java MorfologikSpellerRule.initSpeller + Portuguese.prepareLineForSpeller on plain lists.
	r.InitSpellersFromGetters(language.PortuguesePrepareLineForSpeller, nil)
	// Java static PortugueseTagger / PortugueseSynthesizer.INSTANCE when assets exist.
	if tg := sharedPortugueseSpellerTagger(); tg != nil {
		WirePortugueseSpellerTagger(r, tg)
	}
	if synth := sharedPortugueseSpellerSynth(); synth != nil && synth.BaseSynthesizer != nil {
		base := synth.BaseSynthesizer
		r.SynthesizeForPosTags = func(lemma string, acceptTag func(string) bool) []string {
			return base.SynthesizeForPosTags(lemma, acceptTag)
		}
	}
	return r
}

func NewMorfologikPortugalPortugueseSpellerRule() *MorfologikPortugueseSpellerRule {
	return NewMorfologikPortugueseSpellerRule("pt-PT", PortuguesePTDict, MorfologikPortuguesePTSpellerRuleID)
}

func NewMorfologikBrazilianPortugueseSpellerRule() *MorfologikPortugueseSpellerRule {
	return NewMorfologikPortugueseSpellerRule("pt-BR", PortugueseBRDict, MorfologikPortugueseBRSpellerRuleID)
}

// NewMorfologikMozambiquePortugueseSpellerRule ports pt-MZ (pre-90 / pt-PT-45.dict).
func NewMorfologikMozambiquePortugueseSpellerRule() *MorfologikPortugueseSpellerRule {
	return NewMorfologikPortugueseSpellerRule("pt-MZ", PortuguesePT45Dict, MorfologikPortugueseMZSpellerRuleID)
}

// NewMorfologikAngolaPortugueseSpellerRule ports pt-AO (pt-PT-45.dict).
func NewMorfologikAngolaPortugueseSpellerRule() *MorfologikPortugueseSpellerRule {
	return NewMorfologikPortugueseSpellerRule("pt-AO", PortuguesePT45Dict, MorfologikPortugueseAOSpellerRuleID)
}

// Match ports MorfologikSpellerRule.Match + PT getRuleMatches post-filters.
func (r *MorfologikPortugueseSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.MorfologikSpellerRule == nil {
		return nil, nil
	}
	base, err := r.MorfologikSpellerRule.Match(sentence)
	if err != nil || len(base) == 0 {
		return base, err
	}
	out := make([]*rules.RuleMatch, 0, len(base))
	for _, m := range base {
		if m == nil {
			continue
		}
		word := matchSurface(m, sentence)
		if word == "" {
			out = append(out, m)
			continue
		}
		// Fill suggestions from wired CFSA2 dict when map Speller is empty
		// (Java speller1.getSuggestions always has the binary multi-speller).
		if len(m.GetSuggestedReplacements()) == 0 {
			if sugs := r.wordSuggestions(word); len(sugs) > 0 {
				m.SetSuggestedReplacements(filterDoNotSuggest(sugs))
			}
		}
		// filter do_not_suggest (also via FilterNoSuggestWordsFn on base path).
		// Abbreviations are prepended via GetAdditionalTopSuggestionsFn (Java getAdditionalTopSuggestions),
		// not by replacing the full suggestion list.
		if sugs := m.GetSuggestedReplacements(); len(sugs) > 0 {
			m.SetSuggestedReplacements(filterDoNotSuggest(sugs))
		}

		// clitic verb: drop match when valid
		if r.isValidCliticVerb(word) {
			continue
		}

		// hyphenated word handling
		if strings.Contains(word, "-") {
			parts := strings.Split(word, "-")
			if isTitlecasedHyphenatedWord(parts) {
				// if lower form accepted, drop match (use rule IsMisspelled / FilterDict)
				if !r.wordIsMisspelled(strings.ToLower(word)) {
					continue
				}
			}
			if len(m.GetSuggestedReplacements()) == 0 {
				if ns := r.checkCompoundElements(parts); ns == "" {
					// Java: newSuggestion null → empty rule matches
					continue
				} else {
					m.SetSuggestedReplacements([]string{ns})
				}
			}
		}

		// European 1PL past on BR (Java dialectIssue=true → getIdForDialectIssue)
		if alt := checkEuropeanStyle1PLPastTense(r.VariantCode, word); alt != "" {
			m.Message = "No Brasil, o pretérito perfeito da primeira pessoa do plural escreve-se sem acento."
			m.SetSuggestedReplacements([]string{alt})
			m.SetSpecificRuleId(r.getIdForDialectIssue())
		}

		// diaeresis (Java dialectIssue=false)
		if alt := checkDiaeresis(word); alt != "" {
			m.Message = "No mais recente acordo ortográfico, não se usa mais o trema no português."
			m.SetSuggestedReplacements([]string{alt})
		}

		// dialect surface/lemma map (Java dialectIssue=true)
		if alt := r.dialectAlternative(word); alt != "" {
			other := "europeu"
			if r.VariantCode == "pt-PT" {
				other = "brasileiro"
			}
			m.Message = "Possível erro de ortografia: esta é a grafia utilizada no português " + other + "."
			sug := alt
			if startsWithUppercaseLetter(word) {
				sug = tools.UppercaseFirstChar(alt)
			}
			m.SetSuggestedReplacements([]string{sug})
			m.SetSpecificRuleId(r.getIdForDialectIssue())
		}

		out = append(out, m)
	}
	return out, nil
}

// getIdForDialectIssue ports MorfologikPortugueseSpellerRule.getIdForDialectIssue.
func (r *MorfologikPortugueseSpellerRule) getIdForDialectIssue() string {
	if r == nil {
		return ""
	}
	return r.GetID() + "_DIALECT"
}
