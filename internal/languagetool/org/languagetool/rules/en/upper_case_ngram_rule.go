package en

import (
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

// UpperCaseNgramRule ports org.languagetool.rules.en.UpperCaseNgramRule (simplified).
// Flags mid-sentence Titlecase tokens that look wrong without ngram support;
// when LM is set, Compare scores title vs lower (stub: always prefer lower if not proper).
type UpperCaseNgramRule struct {
	ID string
	LM ngrams.LanguageModel
	// IsException skips known proper nouns / acronyms.
	IsException func(word string) bool
}

func NewUpperCaseNgramRule(lm ngrams.LanguageModel) *UpperCaseNgramRule {
	return &UpperCaseNgramRule{
		ID: "UPPER_CASE_NGRAM_RULE",
		LM: lm,
	}
}

func (r *UpperCaseNgramRule) GetID() string { return r.ID }

// Match flags single-token Titlecase words not at sentence start.
func (r *UpperCaseNgramRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || sentence == nil {
		return nil, nil
	}
	var out []*rules.RuleMatch
	toks := sentence.GetTokensWithoutWhitespace()
	seenContent := false
	for _, tok := range toks {
		if tok == nil {
			continue
		}
		// skip sentence-start markers and first content token
		if tok.IsSentenceStart() {
			continue
		}
		w := tok.GetToken()
		if !seenContent {
			seenContent = true
			continue
		}
		if !isTitleCase(w) || len([]rune(w)) < 2 {
			continue
		}
		if r.IsException != nil && r.IsException(w) {
			continue
		}
		// if all caps acronym (NASA) skip
		if isAllUpper(w) {
			continue
		}
		lower := toLower(w)
		// with LM, could compare; without, always suggest lowercase for mid-sentence titlecase
		if r.LM != nil {
			// prefer lower if its unigram score is higher (simple)
			_ = lower
		}
		m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
			"This word is usually not capitalized in the middle of a sentence.")
		m.SetSuggestedReplacements([]string{lower})
		out = append(out, m)
	}
	return out, nil
}

func isTitleCase(s string) bool {
	rs := []rune(s)
	if len(rs) < 2 {
		return false
	}
	if !unicode.IsUpper(rs[0]) {
		return false
	}
	for _, r := range rs[1:] {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

func isAllUpper(s string) bool {
	has := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			has = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return has
}

func toLower(s string) string {
	return string([]rune(stringToLower(s)))
}

func stringToLower(s string) string {
	b := []rune(s)
	for i, r := range b {
		b[i] = unicode.ToLower(r)
	}
	return string(b)
}

// FirstLongWordToLeftIsUppercase ports the helper used by UpperCaseNgramRule tests:
// looking left from idx, is there a long (>=4 letter) title/upper word before
// reaching a sentence boundary-like marker?
func FirstLongWordToLeftIsUppercase(tokens []*languagetool.AnalyzedTokenReadings, idx int) bool {
	for i := idx - 1; i >= 0; i-- {
		if tokens[i] == nil {
			continue
		}
		if tokens[i].IsSentenceStart() {
			return false
		}
		w := tokens[i].GetToken()
		letters := 0
		for _, r := range w {
			if unicode.IsLetter(r) {
				letters++
			}
		}
		if letters < 4 {
			continue
		}
		return isTitleCase(w) || isAllUpper(w)
	}
	return false
}
