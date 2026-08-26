package el

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikGreekSpellerRuleID ports MorfologikGreekSpellerRule.getId().
	// Java: "MORFOLOGIK_RULE_EL_GR"
	MorfologikGreekSpellerRuleID = "MORFOLOGIK_RULE_EL_GR"
	// GreekSpellerDict ports MorfologikGreekSpellerRule.RESOURCE_FILENAME / getFileName().
	// Java: "/el/hunspell/el_GR.dict"
	GreekSpellerDict = "/el/hunspell/el_GR.dict"
)

// MorfologikGreekSpellerRule ports rules.el.MorfologikGreekSpellerRule.
// Thin pure super: getFileName + getId + isLatinScript()=false; no ignoreToken/filterNoSuggest overrides.
type MorfologikGreekSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

// NewMorfologikGreekSpellerRule ports ctor with null UserConfig / altLanguages.
func NewMorfologikGreekSpellerRule() *MorfologikGreekSpellerRule {
	return NewMorfologikGreekSpellerRuleWithUser(nil)
}

// NewMorfologikGreekSpellerRuleWithUser ports
// MorfologikGreekSpellerRule(messages, language, userConfig, altLanguages).
func NewMorfologikGreekSpellerRuleWithUser(userConfig *languagetool.UserConfig) *MorfologikGreekSpellerRule {
	var userWords []string
	var accepted []string
	if userConfig != nil {
		accepted = userConfig.GetAcceptedWords()
		userWords = morfologik.UserDictWordsForMulti(accepted, userConfig.GetPremiumUid())
	}
	r := &MorfologikGreekSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikGreekSpellerRuleID, "el", GreekSpellerDict, nil),
	}
	// Java isLatinScript() = false → NonLatinScript
	if r.SpellingCheckRule != nil {
		r.NonLatinScript = true
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

// IsLatinScript ports isLatinScript() = false.
func (r *MorfologikGreekSpellerRule) IsLatinScript() bool { return false }
