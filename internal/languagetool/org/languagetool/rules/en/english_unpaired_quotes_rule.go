package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// EnglishUnpairedQuotesRule ports org.languagetool.rules.en.EnglishUnpairedQuotesRule.
// Apostrophe exceptions are POS-only (Java); without tags fail closed to base quote pairing.
type EnglishUnpairedQuotesRule struct {
	*rules.GenericUnpairedQuotesRule
}

func NewEnglishUnpairedQuotesRule(messages map[string]string) *EnglishUnpairedQuotesRule {
	start := []string{"“", "\"", "'", "‘"}
	end := []string{"”", "\"", "'", "’"}
	base := rules.NewGenericUnpairedQuotesRule(messages, start, end)
	base.SetRuleID("EN_UNPAIRED_QUOTES")
	// Java EnglishUnpairedQuotesRule overrides isNotBeginning/EndingApostrophe.
	base.IsNotBeginningApostropheFn = englishIsNotBeginningApostrophe
	base.IsNotEndingApostropheFn = englishIsNotEndingApostrophe
	return &EnglishUnpairedQuotesRule{GenericUnpairedQuotesRule: base}
}

// englishIsNotBeginningApostrophe ports EnglishUnpairedQuotesRule.isNotBeginningApostrophe.
// When POS tags mark contraction/possessive/proper, return false so the apostrophe is not
// treated as an unpaired opening quote (Java).
func englishIsNotBeginningApostrophe(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return true
	}
	if tokens[i].HasPosTag("_apostrophe_contraction_") || tokens[i].HasPosTag("POS") || tokens[i].HasPosTag("NNP") {
		return false
	}
	return true
}

// englishIsNotEndingApostrophe ports EnglishUnpairedQuotesRule.isNotEndingApostrophe.
func englishIsNotEndingApostrophe(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return true
	}
	if tokens[i].HasPosTag("_apostrophe_contraction_") || tokens[i].HasPosTag("POS") || tokens[i].HasPosTag("NNP") {
		return false
	}
	return true
}

func (r *EnglishUnpairedQuotesRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedQuotesRule.MatchList(sentences)
}
