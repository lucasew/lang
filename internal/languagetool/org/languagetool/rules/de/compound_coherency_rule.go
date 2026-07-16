package de

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// CompoundCoherencyRule ports org.languagetool.rules.de.CompoundCoherencyRule
// using surface tokens as lemmas (no German tagger).
type CompoundCoherencyRule struct {
	Messages map[string]string
}

func NewCompoundCoherencyRule(messages map[string]string) *CompoundCoherencyRule {
	return &CompoundCoherencyRule{Messages: messages}
}

func (r *CompoundCoherencyRule) GetID() string { return "DE_COMPOUND_COHERENCY" }

func containsHyphenInside(token string) bool {
	return strings.Contains(token, "-") && !strings.HasPrefix(token, "-") && !strings.HasSuffix(token, "-")
}

func isNumericToken(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (r *CompoundCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	normToText := map[string][]string{}
	pos := 0
	for _, sentence := range sentences {
		for _, atr := range sentence.GetTokensWithoutWhitespace() {
			token := atr.GetToken()
			if token == "" {
				continue
			}
			lemma := token // surface stand-in
			normToken := strings.ToLower(strings.ReplaceAll(lemma, "-", ""))
			if isNumericToken(normToken) {
				break
			}
			if occ, ok := normToText[normToken]; ok {
				foundSame := false
				for _, f := range occ {
					if strings.EqualFold(f, lemma) {
						foundSame = true
						break
					}
				}
				if !foundSame {
					other := occ[0]
					if containsHyphenInside(other) || containsHyphenInside(token) {
						msg := "Uneinheitliche Verwendung von Bindestrichen. Der Text enthält sowohl '" +
							token + "' als auch '" + other + "'."
						rm := rules.NewRuleMatch(r, sentence, pos+atr.GetStartPos(), pos+atr.GetEndPos(), msg)
						if strings.EqualFold(strings.ReplaceAll(token, "-", ""), strings.ReplaceAll(other, "-", "")) {
							rm.SetSuggestedReplacement(other)
						}
						ruleMatches = append(ruleMatches, rm)
					}
				}
			} else {
				normToText[normToken] = []string{lemma}
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
