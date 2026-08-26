package ro

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikRomanianSpellerRuleID ports MorfologikRomanianSpellerRule.getId().
	// Java: "MORFOLOGIK_RULE_RO_RO"
	MorfologikRomanianSpellerRuleID = "MORFOLOGIK_RULE_RO_RO"
	// RomanianSpellerDict ports MorfologikRomanianSpellerRule.RESOURCE_FILENAME / getFileName().
	// Java: "/ro/hunspell/ro_RO.dict"
	RomanianSpellerDict = "/ro/hunspell/ro_RO.dict"
)

// MorfologikRomanianSpellerRule ports rules.ro.MorfologikRomanianSpellerRule.
// Thin pure super: only getFileName + getId; Latin script default (no isLatinScript override).
type MorfologikRomanianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

// NewMorfologikRomanianSpellerRule ports ctor with null UserConfig / altLanguages.
func NewMorfologikRomanianSpellerRule() *MorfologikRomanianSpellerRule {
	return NewMorfologikRomanianSpellerRuleWithUser(nil)
}

// NewMorfologikRomanianSpellerRuleWithUser ports
// MorfologikRomanianSpellerRule(messages, language, userConfig, altLanguages).
func NewMorfologikRomanianSpellerRuleWithUser(userConfig *languagetool.UserConfig) *MorfologikRomanianSpellerRule {
	var userWords []string
	var accepted []string
	if userConfig != nil {
		accepted = userConfig.GetAcceptedWords()
		userWords = morfologik.UserDictWordsForMulti(accepted, userConfig.GetPremiumUid())
	}
	r := &MorfologikRomanianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikRomanianSpellerRuleID, "ro", RomanianSpellerDict, nil),
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
