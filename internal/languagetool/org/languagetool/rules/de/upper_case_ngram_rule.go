package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UpperCaseNgramRule ports org.languagetool.rules.de.UpperCaseNgramRule.
// PseudoProbability(trigram) returns probability of [prev, word, next].
// Without LM hook, Match is empty (Java requires LanguageModel; defaultTempOff).
// No surface heuristic invent when LM is absent.
type UpperCaseNgramRule struct {
	Messages map[string]string
	// Category ports setCategory(CASING).
	Category *rules.Category
	// IssueType ports setLocQualityIssueType(Misspelling).
	IssueType rules.ITSIssueType
	// PseudoProbability ports LanguageModel.getPseudoProbability(list).GetProb().
	// Return 0 for unknown; ratios use these probs.
	PseudoProbability func(trigram []string) float64
	// DefaultTempOff mirrors Java setDefaultTempOff.
	DefaultTempOff bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

const upperCaseNgramThreshold = 50.0

var upperCaseNgramRelevant = map[string]struct{}{
	"tage": {}, "tagen": {},
	"Tage": {}, "Tagen": {},
}

func NewUpperCaseNgramRule(messages map[string]string) *UpperCaseNgramRule {
	r := &UpperCaseNgramRule{
		Messages:       messages,
		Category:       rules.CatCasing.GetCategory(messages),
		IssueType:      rules.ITSMisspelling,
		DefaultTempOff: true,
	}
	// Java: addExamplePair (tagen → Tagen)
	r.AddExamplePair(
		rules.Wrong("Die Suche endete nach 15 <marker>tagen</marker>."),
		rules.Fixed("Die Suche endete nach 15 <marker>Tagen</marker>."),
	)
	return r
}

// NewUpperCaseNgramRuleWithLM ports the Java constructor that requires LanguageModel.
// lm is GetPseudoProbability for a 3-token list (prev, word, next).
func NewUpperCaseNgramRuleWithLM(messages map[string]string, lm func(trigram []string) float64) *UpperCaseNgramRule {
	r := NewUpperCaseNgramRule(messages)
	r.PseudoProbability = lm
	return r
}

func (r *UpperCaseNgramRule) GetID() string { return "DE_UPPER_CASE_NGRAM" }

func (r *UpperCaseNgramRule) GetDescription() string {
	return "Prüft Wörter, ob sie fälschlich groß- oder fälschlich kleingeschrieben sind"
}

func (r *UpperCaseNgramRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *UpperCaseNgramRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSMisspelling
	}
	return r.IssueType
}

// AddExamplePair ports Rule.addExamplePair.
func (r *UpperCaseNgramRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *UpperCaseNgramRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *UpperCaseNgramRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

func (r *UpperCaseNgramRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil || r == nil || r.PseudoProbability == nil {
		// Java always has LM; without it fail closed (no invent).
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		if i+1 >= len(tokens) || tokens[i] == nil || tokens[i-1] == nil || tokens[i+1] == nil {
			continue
		}
		tokenStr := tokens[i].GetToken()
		if _, ok := upperCaseNgramRelevant[tokenStr]; !ok {
			continue
		}
		if tools.IsAllUppercase(tokenStr) {
			continue
		}
		ucToken := tools.UppercaseFirstChar(tokenStr)
		lcToken := tools.LowercaseFirstChar(tokenStr)
		prev, next := tokens[i-1].GetToken(), tokens[i+1].GetToken()

		ucProb := r.PseudoProbability([]string{prev, ucToken, next})
		lcProb := r.PseudoProbability([]string{prev, lcToken, next})
		if ucProb <= 0 {
			ucProb = 1e-20
		}
		if lcProb <= 0 {
			lcProb = 1e-20
		}
		if tools.StartsWithUppercase(tokenStr) {
			ratio := lcProb / ucProb
			if ratio > upperCaseNgramThreshold {
				msg := "Meinten Sie das Verb '" + lcToken + "'? Nur Nomen und Eigennamen werden großgeschrieben."
				rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
				rm.SetSuggestedReplacement(lcToken)
				matches = append(matches, rm)
			}
		} else {
			ratio := ucProb / lcProb
			if ratio > upperCaseNgramThreshold {
				msg := "Meinten Sie das Nomen '" + ucToken + "'? Nomen und Eigennamen werden großgeschrieben."
				rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
				rm.SetSuggestedReplacement(ucToken)
				matches = append(matches, rm)
			}
		}
	}
	return matches
}
