package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikDutchSpellerRuleID ports MorfologikDutchSpellerRule.getId().
	// Java returns "MORFOLOGIK_RULE_NL_NL" (not MORFOLOGIK_RULE_NL).
	MorfologikDutchSpellerRuleID = "MORFOLOGIK_RULE_NL_NL"
	// DutchSpellerDict ports MorfologikDutchSpellerRule.getFileName().
	// Java: "/nl/spelling/nl_NL.dict" (not /nl/hunspell/nl_NL.dict).
	DutchSpellerDict = "/nl/spelling/nl_NL.dict"
	// EnglishIgnorePOS ports the POS tag Java getRuleMatches skips.
	englishIgnorePOS = "_english_ignore_"
)

// MorfologikDutchSpellerRule ports rules.nl.MorfologikDutchSpellerRule.
// ignorePotentiallyMisspelledWord → Dutch.getCompoundAcceptor().acceptCompound.
// getRuleMatches skips tokens tagged _english_ignore_.
// Word lists: /nl/spelling/{ignore,spelling,prohibit}.txt (not hunspell/).
type MorfologikDutchSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikDutchSpellerRule() *MorfologikDutchSpellerRule {
	r := &MorfologikDutchSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikDutchSpellerRuleID, "nl", DutchSpellerDict, nil),
	}
	// Java paths under /nl/spelling/ (not hunspell/).
	if r.SpellingCheckRule != nil {
		r.GetIgnoreFileNameFn = func() string { return "/nl/spelling/ignore.txt" }
		r.GetSpellingFileNameFn = func() string { return "/nl/spelling/spelling.txt" }
		r.GetProhibitFileNameFn = func() string { return "/nl/spelling/prohibit.txt" }
		spelling.ReapplyDefaultSpellingWordLists(r.SpellingCheckRule)
	}
	// Java MorfologikDutchSpellerRule.ignorePotentiallyMisspelledWord:
	// return Dutch.getCompoundAcceptor().acceptCompound(word);
	r.IgnorePotentiallyMisspelledWordFn = func(word string) bool {
		return DefaultCompoundAcceptor.Accept(word)
	}
	// Java getRuleMatches: if tokens[idx].hasPosTag("_english_ignore_") return empty.
	r.SkipTokenFn = func(tok *languagetool.AnalyzedTokenReadings) bool {
		return tok != nil && tok.HasPosTag(englishIgnorePOS)
	}
	return r
}
