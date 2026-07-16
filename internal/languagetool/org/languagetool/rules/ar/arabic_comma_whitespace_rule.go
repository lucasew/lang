package ar

import (
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicCommaWhitespaceRule ports org.languagetool.rules.ar.ArabicCommaWhitespaceRule.
type ArabicCommaWhitespaceRule struct {
	*rules.CommaWhitespaceRule
}

func NewArabicCommaWhitespaceRule(messages map[string]string) *ArabicCommaWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	base.CommaCharacter = "،"
	base.RuleID = "ARABIC_COMMA_PARENTHESIS_WHITESPACE"
	return &ArabicCommaWhitespaceRule{CommaWhitespaceRule: base}
}

func (r *ArabicCommaWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	matches := r.CommaWhitespaceRule.Match(sentence)
	// Tokenizer often glues Arabic comma to words ("هذه،جملة"). Scan raw text too.
	text := sentence.GetText()
	comma := "،"
	seen := map[int]bool{}
	for _, m := range matches {
		seen[m.GetFromPos()] = true
	}
	// missing space after comma: ،X
	for i := 0; i < len(text); {
		idx := strings.Index(text[i:], comma)
		if idx < 0 {
			break
		}
		bytePos := i + idx
		// after comma
		after := bytePos + len(comma)
		if after < len(text) {
			next := text[after]
			if next != ' ' && next != '\n' && next != '\t' && next != 0xC2 { // rough non-space
				// skip if digit (numbers)
				if !isDigitByte(next) {
					from := utf16Offset(text, bytePos)
					to := utf16Offset(text, after)
					// need to include next token char in suggestion range? Java marks comma position
					if !seen[from] {
						// missing space after
						rm := rules.NewRuleMatch(r, sentence, from, to,
							"Put a space after the comma")
						// try to extend to next char for suggestion
						_, size := firstRune(text[after:])
						rm.SetSuggestedReplacement(comma + " " + text[after:after+size])
						// Actually re-match style of CommaWhitespace: from comma to end of next word start
						matches = append(matches, rm)
						seen[from] = true
					}
				}
			}
		}
		// space before comma: X ،
		if bytePos > 0 && text[bytePos-1] == ' ' {
			from := utf16Offset(text, bytePos-1)
			if !seen[from] {
				rm := rules.NewRuleMatch(r, sentence, from, utf16Offset(text, after),
					"Don't put a space before the comma")
				rm.SetSuggestedReplacement(comma)
				matches = append(matches, rm)
				seen[from] = true
			}
		}
		// leading comma at start of token sequence: ،هذه
		if bytePos == 0 || (bytePos > 0 && (text[bytePos-1] == ' ' || text[bytePos-1] == '\n')) {
			// comma after space or start — might also need missing space after if glued
		}
		i = bytePos + len(comma)
	}
	return matches
}

func isDigitByte(b byte) bool { return b >= '0' && b <= '9' }

func utf16Offset(text string, byteIdx int) int {
	n := 0
	for i, r := range text {
		if i >= byteIdx {
			break
		}
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}

func firstRune(s string) (rune, int) {
	for _, r := range s {
		return r, len(string(r))
	}
	return 0, 0
}
