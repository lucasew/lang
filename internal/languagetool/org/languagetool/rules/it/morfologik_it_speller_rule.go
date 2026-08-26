package it

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	// MorfologikItalianSpellerRuleID ports MorfologikItalianSpellerRule.getId().
	// Java returns "MORFOLOGIK_RULE_IT_IT" (not MORFOLOGIK_RULE_IT).
	MorfologikItalianSpellerRuleID = "MORFOLOGIK_RULE_IT_IT"
	// ItalianSpellerDict ports MorfologikItalianSpellerRule.getFileName() → RESOURCE_FILENAME.
	// Java: "/it/hunspell/it_IT.dict"
	ItalianSpellerDict = "/it/hunspell/it_IT.dict"
)

// MorfologikItalianSpellerRule ports rules.it.MorfologikItalianSpellerRule.
// Thin super + orderSuggestions (drop capitalized when lower twin present).
type MorfologikItalianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

// NewMorfologikItalianSpellerRule ports ctor with null UserConfig / altLanguages.
func NewMorfologikItalianSpellerRule() *MorfologikItalianSpellerRule {
	return NewMorfologikItalianSpellerRuleWithUser(nil)
}

// NewMorfologikItalianSpellerRuleWithUser ports
// MorfologikItalianSpellerRule(messages, language, userConfig, altLanguages).
func NewMorfologikItalianSpellerRuleWithUser(userConfig *languagetool.UserConfig) *MorfologikItalianSpellerRule {
	var userWords []string
	var accepted []string
	if userConfig != nil {
		accepted = userConfig.GetAcceptedWords()
		userWords = morfologik.UserDictWordsForMulti(accepted, userConfig.GetPremiumUid())
	}
	r := &MorfologikItalianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikItalianSpellerRuleID, "it", ItalianSpellerDict, nil),
	}
	if userConfig != nil {
		r.SetUserConfig(userConfig)
		// Java SpellingCheckRule: wordsToBeIgnored.addAll(userConfig.getAcceptedWords()) always.
		if r.SpellingCheckRule != nil {
			r.ApplyUserAcceptedWords(accepted)
		}
	}
	// Java MorfologikSpellerRule.initSpeller: binary + spelling.txt + custom + global (+ user FSA if premium).
	r.InitSpellersFromGetters(nil, userWords)
	return r
}

// Match ports parent Match + orderSuggestions (capitalized-dup drop).
// Java: calcSpellerSuggestions → orderSuggestions(SuggestedReplacement list, word).
// Go applies the same filter on final match suggestions (string form of SuggestedReplacement).
func (r *MorfologikItalianSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.MorfologikSpellerRule == nil {
		return nil, nil
	}
	base, err := r.MorfologikSpellerRule.Match(sentence)
	if err != nil || len(base) == 0 {
		return base, err
	}
	for _, m := range base {
		if m == nil {
			continue
		}
		word := matchSurfaceIT(m, sentence)
		sugs := m.GetSuggestedReplacements()
		if len(sugs) == 0 {
			continue
		}
		m.SetSuggestedReplacements(orderItalianSuggestions(sugs, word))
	}
	return base, nil
}

// orderItalianSuggestions ports MorfologikItalianSpellerRule.orderSuggestions:
// drop capitalized suggestion when original word is not capitalized and the
// lowercase form is also present in the suggestion list.
//
// Java walks SuggestedReplacement.getReplacement(); Go uses the string list that
// RuleMatch stores (same surface semantics as base Match).
func orderItalianSuggestions(suggestions []string, word string) []string {
	if len(suggestions) == 0 {
		return suggestions
	}
	// originalSuggestionsStr for membership checks (list contains, not set-only —
	// but set is fine for contains of full replacement strings).
	set := make(map[string]struct{}, len(suggestions))
	for _, s := range suggestions {
		set[s] = struct{}{}
	}
	out := make([]string, 0, len(suggestions))
	for _, sug := range suggestions {
		// !isCapitalizedWord(word) && isCapitalizedWord(sug) && list contains sug.toLowerCase()
		// Java: String.toLowerCase() default locale; Italian Latin → strings.ToLower.
		if !tools.IsCapitalizedWord(word) && tools.IsCapitalizedWord(sug) {
			if _, ok := set[strings.ToLower(sug)]; ok {
				continue
			}
		}
		out = append(out, sug)
	}
	return out
}

func matchSurfaceIT(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
	if m == nil || sent == nil {
		return ""
	}
	// Java RuleMatch FromPos/ToPos are UTF-16 code units (String.substring).
	return rules.UTF16Substring(sent.GetText(), m.GetFromPos(), m.GetToPos())
}
