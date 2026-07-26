package sk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikSlovakSpellerRuleID ports MorfologikSlovakSpellerRule.getId().
	// Java: "MORFOLOGIK_RULE_SK_SK"
	MorfologikSlovakSpellerRuleID = "MORFOLOGIK_RULE_SK_SK"
	// SlovakSpellerDict ports MorfologikSlovakSpellerRule.RESOURCE_FILENAME / getFileName().
	// Java: "/sk/hunspell/sk_SK.dict"
	SlovakSpellerDict = "/sk/hunspell/sk_SK.dict"
)

// MorfologikSlovakSpellerRule ports rules.sk.MorfologikSlovakSpellerRule.
// Thin pure super: only getFileName + getId; no ignoreToken / filterNoSuggest overrides.
type MorfologikSlovakSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

// NewMorfologikSlovakSpellerRule ports ctor with null UserConfig / altLanguages.
func NewMorfologikSlovakSpellerRule() *MorfologikSlovakSpellerRule {
	return NewMorfologikSlovakSpellerRuleWithUser(nil)
}

// NewMorfologikSlovakSpellerRuleWithUser ports
// MorfologikSlovakSpellerRule(messages, language, userConfig, altLanguages).
func NewMorfologikSlovakSpellerRuleWithUser(userConfig *languagetool.UserConfig) *MorfologikSlovakSpellerRule {
	var userWords []string
	var accepted []string
	if userConfig != nil {
		accepted = userConfig.GetAcceptedWords()
		userWords = morfologik.UserDictWordsForMulti(accepted, userConfig.GetPremiumUid())
	}
	r := &MorfologikSlovakSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikSlovakSpellerRuleID, "sk", SlovakSpellerDict, nil),
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
