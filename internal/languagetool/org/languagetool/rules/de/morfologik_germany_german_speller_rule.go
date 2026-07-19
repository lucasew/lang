package de

import (
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
}

// NewMorfologikGermanyGermanSpellerRule ports the Java constructor.
// messages is accepted for call-site parity; unused until ResourceBundle twin exists.
func NewMorfologikGermanyGermanSpellerRule(messages map[string]string) *MorfologikGermanyGermanSpellerRule {
	_ = messages
	return &MorfologikGermanyGermanSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikGermanyGermanSpellerRuleID, "de", MorfologikGermanyGermanDict, nil),
	}
}

// GetMorfologikDictFilename is a Go helper for tests/callers (Java getFileName).
func (r *MorfologikGermanyGermanSpellerRule) GetMorfologikDictFilename() string {
	if r == nil {
		return MorfologikGermanyGermanDict
	}
	return r.GetFileName()
}
