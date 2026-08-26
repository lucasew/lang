package de

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// Match skip helpers: URL/email/immunized/quoted compound (HunspellRule + German).

// reStartsWithTwoUppercase ports HunspellRule.STARTS_WITH_TWO_UPPERCASE_CHARS
// for high-confidence suggestions (e.g. HAus → Haus).
var reStartsWithTwoUppercase = regexp.MustCompile(`^[A-Z][A-Z]\p{Ll}+$`)

// shouldSkipSpellToken ports HunspellRule skip conditions for a non-whitespace token:
// immunized, ignored-by-speller, URL, email, quoted compound, _english_ignore_.
func (r *GermanSpellerRule) shouldSkipSpellToken(sentence *languagetool.AnalyzedSentence, tokens []*languagetool.AnalyzedTokenReadings, idx int) bool {
	if r == nil || idx < 0 || idx >= len(tokens) || tokens[idx] == nil {
		return true
	}
	tok := tokens[idx]
	word := tok.GetToken()
	if tok.IsImmunized() || tok.IsIgnoredBySpeller() {
		return true
	}
	if spelling.IsUrl(word) || spelling.IsEMail(word) || tokenizers.IsURL(word) || tokenizers.IsEMail(word) {
		return true
	}
	if tok.HasPosTag("_english_ignore_") {
		return true
	}
	if r.isQuotedCompoundNonBlank(tokens, idx, word) {
		return true
	}
	_ = sentence
	return false
}

// isQuotedCompoundNonBlank ports GermanSpellerRule.isQuotedCompound for non-blank token lists:
// token starts with "-" and neighbors form „Word“- / "Word"- style quotes.
// With non-blank tokens: [„, Word, “, -Magazin] → idx of -Magazin is 3.
func (r *GermanSpellerRule) isQuotedCompoundNonBlank(tokens []*languagetool.AnalyzedTokenReadings, idx int, token string) bool {
	if idx < 3 || !strings.HasPrefix(token, "-") {
		return false
	}
	closeQ := tokens[idx-1]
	openQ := tokens[idx-3]
	if closeQ == nil || openQ == nil {
		return false
	}
	c, o := closeQ.GetToken(), openQ.GetToken()
	closing := c == "“" || c == "\"" || c == "»" || c == "›"
	opening := o == "„" || o == "\"" || o == "«" || o == "‹"
	return closing && opening
}

// isFirstItemHighConfidenceSuggestion ports HunspellRule.isFirstItemHighConfidenceSuggestion for DE.
func (r *GermanSpellerRule) isFirstItemHighConfidenceSuggestion(word string, sugs []string) bool {
	if len(sugs) == 0 || word == "IPs" {
		return false
	}
	if !strings.EqualFold(word, sugs[0]) {
		return false
	}
	if !reStartsWithTwoUppercase.MatchString(word) {
		return false
	}
	if strings.HasSuffix(word, "s") && isAllUpperLetters(sugs[0]) {
		return false
	}
	return true
}

func isAllUpperLetters(s string) bool {
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
