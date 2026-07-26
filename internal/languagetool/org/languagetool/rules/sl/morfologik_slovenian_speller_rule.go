package sl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikSlovenianSpellerRuleID ports MorfologikSlovenianSpellerRule.getId().
	// Java: "MORFOLOGIK_RULE_SL_SI"
	MorfologikSlovenianSpellerRuleID = "MORFOLOGIK_RULE_SL_SI"
	// SlovenianSpellerDict ports MorfologikSlovenianSpellerRule.RESOURCE_FILENAME / getFileName().
	// Java: "/sl/hunspell/sl_SI.dict"
	SlovenianSpellerDict = "/sl/hunspell/sl_SI.dict"
)

// MorfologikSlovenianSpellerRule ports rules.sl.MorfologikSlovenianSpellerRule.
// Thin pure super: only getFileName + getId; no ignoreToken / filterNoSuggest overrides.
type MorfologikSlovenianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

// NewMorfologikSlovenianSpellerRule ports ctor with null UserConfig / altLanguages.
func NewMorfologikSlovenianSpellerRule() *MorfologikSlovenianSpellerRule {
	return NewMorfologikSlovenianSpellerRuleWithUser(nil)
}

// NewMorfologikSlovenianSpellerRuleWithUser ports
// MorfologikSlovenianSpellerRule(messages, language, userConfig, altLanguages).
func NewMorfologikSlovenianSpellerRuleWithUser(userConfig *languagetool.UserConfig) *MorfologikSlovenianSpellerRule {
	var userWords []string
	var accepted []string
	if userConfig != nil {
		accepted = userConfig.GetAcceptedWords()
		userWords = morfologik.UserDictWordsForMulti(accepted, userConfig.GetPremiumUid())
	}
	r := &MorfologikSlovenianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikSlovenianSpellerRuleID, "sl", SlovenianSpellerDict, nil),
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
