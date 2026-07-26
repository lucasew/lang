package ekavian

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikEkavianSpellerRuleID ports MorfologikEkavianSpellerRule.RULE_ID / getId().
	// Java: "MORFOLOGIK_RULE_SR_EKAVIAN"
	MorfologikEkavianSpellerRuleID = "MORFOLOGIK_RULE_SR_EKAVIAN"
	// EkavianSpellerDict ports getFileName().
	// Java BASE_DICTIONARY_PATH + "serbian.dict" → "/sr/dictionary/ekavian/serbian.dict"
	EkavianSpellerDict = "/sr/dictionary/ekavian/serbian.dict"
	// Java path overrides under BASE_DICTIONARY_PATH (not hunspell/).
	EkavianIgnoreFile   = "sr/dictionary/ekavian/ignored.txt"
	EkavianSpellingFile = "sr/dictionary/ekavian/spelling.txt"
	EkavianProhibitFile = "sr/dictionary/ekavian/prohibit.txt"
)

// MorfologikEkavianSpellerRule ports rules.sr.ekavian.MorfologikEkavianSpellerRule.
// Thin super with path overrides + example pair; Java does not override isLatinScript
// (default true → pure non-Latin tokens ignored via pHasNoLetterLatin).
type MorfologikEkavianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	// incorrectExamples / correctExamples port Rule.addExamplePair (not on SpellingCheckRule).
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// NewMorfologikEkavianSpellerRule ports ctor with null UserConfig / empty altLanguages.
func NewMorfologikEkavianSpellerRule() *MorfologikEkavianSpellerRule {
	return NewMorfologikEkavianSpellerRuleWithUser(nil)
}

// NewMorfologikEkavianSpellerRuleWithUser ports
// MorfologikEkavianSpellerRule(messages, language, userConfig, altLanguages).
func NewMorfologikEkavianSpellerRuleWithUser(userConfig *languagetool.UserConfig) *MorfologikEkavianSpellerRule {
	var userWords []string
	var accepted []string
	if userConfig != nil {
		accepted = userConfig.GetAcceptedWords()
		userWords = morfologik.UserDictWordsForMulti(accepted, userConfig.GetPremiumUid())
	}
	r := &MorfologikEkavianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikEkavianSpellerRuleID, "sr", EkavianSpellerDict, nil),
	}
	// Java getIgnoreFileName / getSpellingFileName / getProhibitFileName under dictionary/ekavian/.
	if r.SpellingCheckRule != nil {
		r.GetIgnoreFileNameFn = func() string { return "/" + EkavianIgnoreFile }
		r.GetSpellingFileNameFn = func() string { return "/" + EkavianSpellingFile }
		r.GetProhibitFileNameFn = func() string { return "/" + EkavianProhibitFile }
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
	r.InitSpellersFromGetters(nil, userWords)
	// Java: бткие → битке
	r.AddExamplePair(
		rules.Wrong("Изгубила све сам <marker>бткие</marker>, ал' још водим рат."),
		rules.Fixed("Изгубила све сам <marker>битке</marker>, ал' још водим рат."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *MorfologikEkavianSpellerRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *MorfologikEkavianSpellerRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *MorfologikEkavianSpellerRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// GetIgnoreFileName ports getIgnoreFileName.
func (r *MorfologikEkavianSpellerRule) GetIgnoreFileName() string { return "/" + EkavianIgnoreFile }

// GetSpellingFileName ports getSpellingFileName.
func (r *MorfologikEkavianSpellerRule) GetSpellingFileName() string {
	return "/" + EkavianSpellingFile
}

// GetProhibitFileName ports getProhibitFileName.
func (r *MorfologikEkavianSpellerRule) GetProhibitFileName() string {
	return "/" + EkavianProhibitFile
}
