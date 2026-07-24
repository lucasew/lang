package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikGermanyGermanSpellerRuleID ports MorfologikGermanyGermanSpellerRule.getId().
	// Java: "MORFOLOGIK_RULE_DE_DE" (not GERMAN_SPELLER_RULE / GermanSpellerRule).
	MorfologikGermanyGermanSpellerRuleID = "MORFOLOGIK_RULE_DE_DE"
	// MorfologikGermanyGermanDict ports getFileName() → RESOURCE_FILENAME.
	// Java: "/de/hunspell/de_DE.dict"
	MorfologikGermanyGermanDict = "/de/hunspell/de_DE.dict"
)

// MorfologikGermanyGermanSpellerRule ports
// org.languagetool.rules.de.MorfologikGermanyGermanSpellerRule
// (deprecated non-compound Morfologik speller; distinct from GermanSpellerRule).
//
// Not an alias of GermanSpellerRule — Java extends MorfologikSpellerRule directly.
type MorfologikGermanyGermanSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// NewMorfologikGermanyGermanSpellerRule ports the Java constructor.
// messages is accepted for call-site parity; unused until ResourceBundle twin exists.
func NewMorfologikGermanyGermanSpellerRule(messages map[string]string) *MorfologikGermanyGermanSpellerRule {
	_ = messages
	r := &MorfologikGermanyGermanSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikGermanyGermanSpellerRuleID, "de", MorfologikGermanyGermanDict, nil),
	}
	// Java MorfologikSpellerRule.initSpeller uses Language.prepareLineForSpeller
	// (German E/S/N flags) — not ExpandingReader (that is GermanSpellerRule.getSpeller only).
	r.InitSpellersFromGetters(language.GermanPrepareLineForSpeller, nil)
	// Java: nromale → normale
	r.AddExamplePair(
		rules.Wrong("LanguageTool kann mehr als eine <marker>nromale</marker> Rechtschreibprüfung."),
		rules.Fixed("LanguageTool kann mehr als eine <marker>normale</marker> Rechtschreibprüfung."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *MorfologikGermanyGermanSpellerRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *MorfologikGermanyGermanSpellerRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil {
		return nil
	}
	return r.incorrectExamples
}

// GetMorfologikDictFilename is a Go helper for tests/callers (Java getFileName).
func (r *MorfologikGermanyGermanSpellerRule) GetMorfologikDictFilename() string {
	if r == nil {
		return MorfologikGermanyGermanDict
	}
	return r.GetFileName()
}
