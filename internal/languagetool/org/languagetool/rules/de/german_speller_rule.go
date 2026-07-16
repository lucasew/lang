package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanSpellerRule is a stand-in for GermanSpellerRule until Morfologik/hunspell
// dictionaries are wired. Match is a no-op (no false positives without a dict).
type GermanSpellerRule struct {
	Messages map[string]string
	// LanguageVariant: "", "AT", "CH"
	LanguageVariant string
}

func NewGermanSpellerRule(messages map[string]string) *GermanSpellerRule {
	return &GermanSpellerRule{Messages: messages}
}

func NewAustrianGermanSpellerRule(messages map[string]string) *GermanSpellerRule {
	return &GermanSpellerRule{Messages: messages, LanguageVariant: "AT"}
}

func NewSwissGermanSpellerRule(messages map[string]string) *GermanSpellerRule {
	return &GermanSpellerRule{Messages: messages, LanguageVariant: "CH"}
}

func NewMorfologikGermanyGermanSpellerRule(messages map[string]string) *GermanSpellerRule {
	// same soft stand-in; ID differs for registry wiring
	return &GermanSpellerRule{Messages: messages}
}

func (r *GermanSpellerRule) GetID() string {
	switch r.LanguageVariant {
	case "AT":
		return "AUSTRIAN_GERMAN_SPELLER_RULE"
	case "CH":
		return "SWISS_GERMAN_SPELLER_RULE"
	default:
		return "GERMAN_SPELLER_RULE"
	}
}

// GetMessage returns a generic misspelling message (Java builds richer text).
func (r *GermanSpellerRule) GetMessage(word, suggestion string) string {
	if suggestion == "" {
		return "Möglicher Rechtschreibfehler: " + word
	}
	return "Möglicher Rechtschreibfehler: " + word + " → " + suggestion
}

func (r *GermanSpellerRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	// Full spelling needs dictionary backends.
	return nil
}

// IsMisspelled is unknown without a dictionary; treat as not misspelled (fail open for style).
func (r *GermanSpellerRule) IsMisspelled(word string) bool {
	return false
}
