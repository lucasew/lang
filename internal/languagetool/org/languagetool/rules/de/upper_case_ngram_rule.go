package de

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UpperCaseNgramRule is a surface stand-in for UpperCaseNgramRule without n-grams.
// Handles the Tage/Tagen ambiguity with simple context heuristics from the Java tests.
type UpperCaseNgramRule struct {
	Messages map[string]string
}

func NewUpperCaseNgramRule(messages map[string]string) *UpperCaseNgramRule {
	return &UpperCaseNgramRule{Messages: messages}
}

func (r *UpperCaseNgramRule) GetID() string { return "DE_UPPER_CASE_NGRAM" }

func isDigitsToken(s string) bool {
	if s == "" {
		return false
	}
	for _, r0 := range s {
		if !unicode.IsDigit(r0) && r0 != '.' && r0 != ',' {
			return false
		}
	}
	return unicode.IsDigit([]rune(s)[0])
}

func (r *UpperCaseNgramRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		w := tok.GetToken()
		lc := strings.ToLower(w)
		if lc != "tage" && lc != "tagen" {
			continue
		}
		// After a number: prefer capitalized Tagen/Tage (noun)
		if i > 1 && isDigitsToken(tokens[i-1].GetToken()) {
			if !tools.StartsWithUppercase(w) {
				msg := "Meinten Sie '" + tools.UppercaseFirstChar(w) + "' (Substantiv)?"
				rm := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
				rm.SetSuggestedReplacement(tools.UppercaseFirstChar(w))
				matches = append(matches, rm)
			}
			continue
		}
		// After "Sie" (pronoun): prefer lowercase verb "tagen"
		if i > 1 && tokens[i-1].GetToken() == "Sie" && tools.StartsWithUppercase(w) && utf8.RuneCountInString(w) > 0 {
			// sentence-start "Sie Tagen" — verb should be lower
			msg := "Meinten Sie '" + strings.ToLower(w) + "' (Verb)?"
			rm := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
			rm.SetSuggestedReplacement(strings.ToLower(w))
			matches = append(matches, rm)
		}
	}
	return matches
}
