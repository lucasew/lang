package jekavian

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikJekavianSpellerRuleID ports MorfologikJekavianSpellerRule.RULE_ID / getId().
	// Java: "MORFOLOGIK_RULE_SR_JEKAVIAN"
	MorfologikJekavianSpellerRuleID = "MORFOLOGIK_RULE_SR_JEKAVIAN"
	// JekavianSpellerDict ports getFileName().
	// Java BASE_DICTIONARY_PATH + "serbian.dict" → "/sr/dictionary/jekavian/serbian.dict"
	JekavianSpellerDict = "/sr/dictionary/jekavian/serbian.dict"
	// Java path overrides under BASE_DICTIONARY_PATH (not hunspell/).
	JekavianIgnoreFile   = "sr/dictionary/jekavian/ignored.txt"
	JekavianSpellingFile = "sr/dictionary/jekavian/spelling.txt"
	JekavianProhibitFile = "sr/dictionary/jekavian/prohibit.txt"
)

// MorfologikJekavianSpellerRule ports rules.sr.jekavian.MorfologikJekavianSpellerRule.
// Thin super with path overrides; Java example pair is commented out (do not invent).
// Java does not override isLatinScript (default true → pure non-Latin tokens ignored
// via pHasNoLetterLatin).
type MorfologikJekavianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

// NewMorfologikJekavianSpellerRule ports ctor with null UserConfig / empty altLanguages.
func NewMorfologikJekavianSpellerRule() *MorfologikJekavianSpellerRule {
	return NewMorfologikJekavianSpellerRuleWithUser(nil)
}

// NewMorfologikJekavianSpellerRuleWithUser ports
// MorfologikJekavianSpellerRule(messages, language, userConfig, altLanguages).
func NewMorfologikJekavianSpellerRuleWithUser(userConfig *languagetool.UserConfig) *MorfologikJekavianSpellerRule {
	var userWords []string
	var accepted []string
	if userConfig != nil {
		accepted = userConfig.GetAcceptedWords()
		userWords = morfologik.UserDictWordsForMulti(accepted, userConfig.GetPremiumUid())
	}
	r := &MorfologikJekavianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikJekavianSpellerRuleID, "sr", JekavianSpellerDict, nil),
	}
	// Java getIgnoreFileName / getSpellingFileName / getProhibitFileName under dictionary/jekavian/.
	if r.SpellingCheckRule != nil {
		r.GetIgnoreFileNameFn = func() string { return "/" + JekavianIgnoreFile }
		r.GetSpellingFileNameFn = func() string { return "/" + JekavianSpellingFile }
		r.GetProhibitFileNameFn = func() string { return "/" + JekavianProhibitFile }
		// Java does not override getAdditional*FileNames — keep SpellingCheckRule defaults
		// (short+custom + spelling_global / prohibit_custom; missing files fail-closed).
		spelling.ReapplyDefaultSpellingWordLists(r.SpellingCheckRule)
		// Java SpellingCheckRule: wordsToBeIgnored.addAll(userConfig.getAcceptedWords()) always.
		if userConfig != nil {
			r.ApplyUserAcceptedWords(accepted)
		}
	}
	if userConfig != nil {
		r.SetUserConfig(userConfig)
	}
	// Java MorfologikSpellerRule.initSpeller: binary + spelling.txt + custom + global (+ user FSA if premium).
	// Java example pair is commented out — do not call AddExamplePair.
	r.InitSpellersFromGetters(nil, userWords)
	return r
}

// GetIgnoreFileName ports getIgnoreFileName.
func (r *MorfologikJekavianSpellerRule) GetIgnoreFileName() string { return "/" + JekavianIgnoreFile }

// GetSpellingFileName ports getSpellingFileName.
func (r *MorfologikJekavianSpellerRule) GetSpellingFileName() string {
	return "/" + JekavianSpellingFile
}

// GetProhibitFileName ports getProhibitFileName.
func (r *MorfologikJekavianSpellerRule) GetProhibitFileName() string {
	return "/" + JekavianProhibitFile
}
