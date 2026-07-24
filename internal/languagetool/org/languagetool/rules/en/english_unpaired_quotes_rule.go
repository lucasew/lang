package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// EnglishUnpairedQuotesRule ports org.languagetool.rules.en.EnglishUnpairedQuotesRule.
//
// Java: extends GenericUnpairedQuotesRule with EN quote symbol pairs, id EN_UNPAIRED_QUOTES,
// insights URL, one example pair, and POS-only overrides of isNotBeginning/EndingApostrophe
// (does not call super — only _apostrophe_contraction_ / POS / NNP skip quote pairing).
// Constructor Language parameter is unused in Java.
type EnglishUnpairedQuotesRule struct {
	*rules.GenericUnpairedQuotesRule
}

// NewEnglishUnpairedQuotesRule ports EnglishUnpairedQuotesRule(messages, language).
// language is unused in Java; omitted here.
func NewEnglishUnpairedQuotesRule(messages map[string]string) *EnglishUnpairedQuotesRule {
	// Java EN_START_SYMBOLS / EN_END_SYMBOLS (order differs from Generic defaults).
	start := []string{"“", "\"", "'", "‘"}
	end := []string{"”", "\"", "'", "’"}
	base := rules.NewGenericUnpairedQuotesRule(messages, start, end)
	base.SetRuleID("EN_UNPAIRED_QUOTES")
	// Java setUrl(Tools.getUrl("…#what-are-quotation-marks")); addExamplePair(...)
	base.URL = "https://languagetool.org/insights/post/punctuation-guide/#what-are-quotation-marks"
	base.AddExamplePair(
		rules.Wrong("\"I'm over here,<marker></marker> she said."),
		rules.Fixed("\"I'm over here,<marker>\"</marker> she said."),
	)
	// Java overrides isNotBeginningApostrophe / isNotEndingApostrophe (no super call).
	base.IsNotBeginningApostropheFn = englishIsNotBeginningApostrophe
	base.IsNotEndingApostropheFn = englishIsNotEndingApostrophe
	return &EnglishUnpairedQuotesRule{GenericUnpairedQuotesRule: base}
}

// englishIsNotBeginningApostrophe ports EnglishUnpairedQuotesRule.isNotBeginningApostrophe.
// return false → treated as contraction/possessive apostrophe (skip quote pairing).
// return true  → treat token as a real quote (Java default when POS tags absent).
func englishIsNotBeginningApostrophe(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return true
	}
	if tokens[i].HasPosTag("_apostrophe_contraction_") || tokens[i].HasPosTag("POS") || tokens[i].HasPosTag("NNP") {
		return false
	}
	return true
}

// englishIsNotEndingApostrophe ports EnglishUnpairedQuotesRule.isNotEndingApostrophe
// (identical POS gates to isNotBeginningApostrophe in Java).
func englishIsNotEndingApostrophe(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return true
	}
	if tokens[i].HasPosTag("_apostrophe_contraction_") || tokens[i].HasPosTag("POS") || tokens[i].HasPosTag("NNP") {
		return false
	}
	return true
}

// MatchList ports TextLevelRule match(List) via GenericUnpairedQuotesRule.
func (r *EnglishUnpairedQuotesRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedQuotesRule.MatchList(sentences)
}
