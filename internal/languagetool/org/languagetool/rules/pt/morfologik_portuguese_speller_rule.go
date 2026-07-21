package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Portuguese speller variants — IDs and dict paths from
// org.languagetool.rules.pt.MorfologikPortugueseSpellerRule.
const (
	// getId(): "MORFOLOGIK_RULE_" + shortCodeWithCountryAndVariant uppercased.
	MorfologikPortuguesePTSpellerRuleID = "MORFOLOGIK_RULE_PT_PT"
	MorfologikPortugueseBRSpellerRuleID = "MORFOLOGIK_RULE_PT_BR"
	// Java: dictFilepath = "/pt/spelling/" + getDictFilename() + ".dict"
	// pt-PT → pt-PT-90; pt-BR → pt-BR (not /pt/hunspell/…).
	PortuguesePTDict = "/pt/spelling/pt-PT-90.dict"
	PortugueseBRDict = "/pt/spelling/pt-BR.dict"
	// EnglishIgnorePOS ports POS tag skipped in Java getRuleMatches.
	englishIgnorePOS = "_english_ignore_"
)

// MorfologikPortugueseSpellerRule ports rules.pt.MorfologikPortugueseSpellerRule.
// Word lists: pt/ignore.txt, pt/prohibit.txt, pt/spelling.txt, pt/multiwords.txt.
// getRuleMatches post-filters: _english_ignore_, titlecase-hyphen, diaeresis,
// European 1PL past (pt-BR), dialect surface map, compound-element suggestions,
// do_not_suggest filter. Clitic-verb needs TagPOS (fail-closed when unset).
type MorfologikPortugueseSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	VariantCode string
	// dialectMap ports dialectAlternationMapping for this variant.
	dialectMap map[string]string
	// TagPOS optional PortugueseTagger.tag surface → POS tags (clitic verb).
	TagPOS func(word string) []string
	// TagLemma optional lemmas for clitic dialect invalidation.
	TagLemma func(word string) []string
}

func NewMorfologikPortugueseSpellerRule(variantCode, dict, id string) *MorfologikPortugueseSpellerRule {
	r := &MorfologikPortugueseSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(id, "pt", dict, nil),
		VariantCode:           variantCode,
		dialectMap:            loadDialectAlternationMapping(variantCode),
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
	}
	// Java getRuleMatches: if tokens[idx].hasPosTag("_english_ignore_") return empty.
	r.SkipTokenFn = func(tok *languagetool.AnalyzedTokenReadings) bool {
		return tok != nil && tok.HasPosTag(englishIgnorePOS)
	}
	// Java MorfologikSpellerRule.initSpeller + Portuguese.prepareLineForSpeller on plain lists.
	r.InitSpellersFromGetters(language.PortuguesePrepareLineForSpeller, nil)
	return r
}

func NewMorfologikPortugalPortugueseSpellerRule() *MorfologikPortugueseSpellerRule {
	return NewMorfologikPortugueseSpellerRule("pt-PT", PortuguesePTDict, MorfologikPortuguesePTSpellerRuleID)
}

func NewMorfologikBrazilianPortugueseSpellerRule() *MorfologikPortugueseSpellerRule {
	return NewMorfologikPortugueseSpellerRule("pt-BR", PortugueseBRDict, MorfologikPortugueseBRSpellerRuleID)
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
				m.SetSuggestedReplacements(sugs)
			}
		}
		// filter do_not_suggest from any existing suggestions
		if sugs := m.GetSuggestedReplacements(); len(sugs) > 0 {
			m.SetSuggestedReplacements(filterDoNotSuggest(sugs))
		}
		// abbreviation top suggestion
		if isAbbreviation(word) {
			m.SetSuggestedReplacements([]string{word + "."})
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

		// dialect surface map (Java dialectIssue=true)
		if alt := r.dialectAlternativeSurface(word); alt != "" {
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
