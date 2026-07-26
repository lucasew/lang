package uk

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UkrainianUppercaseSentenceStartRule ports
// org.languagetool.rules.uk.UkrainianUppercaseSentenceStartRule.
// Extends UppercaseSentenceStartRule; overrides isException for list items а) б) в).
type UkrainianUppercaseSentenceStartRule struct {
	*rules.UppercaseSentenceStartRule
}

// Java: private static final Pattern PATTERN = Pattern.compile("[а-яіїєґ]");
// Matcher.matches() ⇒ full string match (single Ukrainian lowercase letter).
var ukUppercaseListLetter = regexp.MustCompile(`^[а-яіїєґ]$`)

// NewUkrainianUppercaseSentenceStartRule ports
// UkrainianUppercaseSentenceStartRule(ResourceBundle, Language, IncorrectExample, CorrectExample).
// messages map is the ResourceBundle twin; language short code is "uk".
func NewUkrainianUppercaseSentenceStartRule(messages map[string]string) *UkrainianUppercaseSentenceStartRule {
	base := rules.NewUppercaseSentenceStartRule(messages, "uk")
	base.IsException = func(tokens []*languagetool.AnalyzedTokenReadings, tokenIdx int) bool {
		// Java:
		// if (tokenIdx == 1 && tokenIdx < tokens.length-1
		//     && PATTERN.matcher(tokens[tokenIdx].getCleanToken()).matches()
		//     && tokens[tokenIdx+1].getToken().equals(")")) return true;
		if tokenIdx == 1 && tokenIdx < len(tokens)-1 &&
			ukUppercaseListLetter.MatchString(tokens[tokenIdx].GetCleanToken()) &&
			tokens[tokenIdx+1].GetToken() == ")" {
			return true
		}
		return false
	}
	return &UkrainianUppercaseSentenceStartRule{UppercaseSentenceStartRule: base}
}
