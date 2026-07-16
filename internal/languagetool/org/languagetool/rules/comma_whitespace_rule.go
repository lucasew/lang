package rules

import (
	"regexp"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CommaWhitespaceRule ports org.languagetool.rules.CommaWhitespaceRule.
type CommaWhitespaceRule struct {
	Messages              map[string]string
	QuotesWhitespaceCheck bool
	fileExt               *regexp.Regexp
	domain                *regexp.Regexp
}

func NewCommaWhitespaceRule(messages map[string]string) *CommaWhitespaceRule {
	return &CommaWhitespaceRule{
		Messages:              messages,
		QuotesWhitespaceCheck: true,
		fileExt:               regexp.MustCompile(`^([a-z]{3,4}|[A-Z]{3,4}|ai|mp[34]|MP[34])(-.+)?$`),
		domain:                regexp.MustCompile(`(?i)^(com|org|net|int|edu|gov|mil|[a-z]{2})$`),
	}
}

func (r *CommaWhitespaceRule) GetID() string           { return "COMMA_PARENTHESIS_WHITESPACE" }
func (r *CommaWhitespaceRule) GetCommaCharacter() string { return "," }

func (r *CommaWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokens()
	prevToken, prevPrevToken := "", ""
	prevWhite := false
	for i := 0; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		isWhitespace := isWhitespaceToken(tokens[i])
		twoSuggestions := false
		var msg, suggestionText string
		msgSet := false

		if isWhitespace && isLeftBracket(prevToken) {
			isException := i+1 < len(tokens) && prevToken == "[" && token == " " && tokens[i+1].GetToken() == "]"
			if !isException {
				msg = r.msg("no_space_after", "Don't put a space after this")
				suggestionText = prevToken
				msgSet = true
			}
		} else if isWhitespace && isQuote(prevToken) && r.QuotesWhitespaceCheck && prevPrevToken == " " {
			msg = r.msg("no_space_around_quotes", "Don't put spaces around quotes")
			suggestionText = prevToken
			twoSuggestions = true
			msgSet = true
		} else if !isWhitespace && prevToken == r.GetCommaCharacter() &&
			!isQuote(token) && !isHyphenOrComma(token) &&
			!containsDigit(prevPrevToken) && !containsDigit(token) && prevPrevToken != "," {
			msg = r.msg("missing_space_after_comma", "Put a space after the comma")
			suggestionText = r.GetCommaCharacter() + " " + tokens[i].GetToken()
			msgSet = true
		} else if prevWhite {
			if isRightBracket(token) {
				isException := token == "]" && prevToken == " " && prevPrevToken == "["
				if !isException {
					msg = r.msg("no_space_before", "Don't put a space before this")
					suggestionText = token
					msgSet = true
				}
			} else if token == r.GetCommaCharacter() {
				msg = r.msg("space_after_comma", "Don't put a space before the comma")
				suggestionText = r.GetCommaCharacter()
				msgSet = true
				if i+1 < len(tokens) && tokens[i+1].GetToken() == r.GetCommaCharacter() {
					msgSet = false
					msg = ""
				}
				if i+1 < len(tokens) && !tokens[i+1].IsWhitespace() {
					suggestionText = r.GetCommaCharacter() + " "
				}
			} else if token == "." && !r.isDomain(tokens, i+1) && !r.isFileExtension(tokens, i+1) {
				msg = r.msg("no_space_before_dot", "Don't put a space before the period")
				suggestionText = "."
				msgSet = true
				if i+1 < len(tokens) && isDigitOrDot(tokens[i+1].GetToken()) {
					msgSet = false
					msg = ""
				} else if i+2 < len(tokens) && tokens[i+1].GetToken() == "/" {
					// ./validate.sh
					next := tokens[i+2].GetToken()
					ok, _ := regexp.MatchString(`^[a-zA-Z]+$`, next)
					if ok {
						msgSet = false
						msg = ""
					}
				}
			}
		}

		if msgSet && msg != "" && !tokens[i].IsImmunized() {
			fromPos := tokens[i-1].GetStartPos()
			if twoSuggestions {
				fromPos = tokens[i-2].GetStartPos()
			}
			toPos := tokens[i].GetEndPos()
			text := sentence.GetText()
			if toPos <= len(text) && toPos <= utf16Len(text) {
				// substring by UTF-16 indices is complex; use rune approx for BMP tests
				// GetText returns original; compare marked region via byte for BMP
			}
			// Java: text.substring(fromPos, toPos) with UTF-16 indices
			marked := utf16Substring(text, fromPos, toPos)
			if marked == suggestionText && !twoSuggestions {
				prevPrevToken = prevToken
				prevToken = token
				prevWhite = isWhitespace && !tokens[i].IsFieldCode()
				continue
			}
			rm := NewRuleMatch(r, sentence, fromPos, toPos, msg)
			if twoSuggestions {
				rm.SuggestedReplacements = []string{suggestionText + " ", " " + suggestionText}
			} else {
				rm.SetSuggestedReplacement(suggestionText)
			}
			ruleMatches = append(ruleMatches, rm)
		}
		prevPrevToken = prevToken
		prevToken = token
		prevWhite = isWhitespace && !tokens[i].IsFieldCode()
	}
	return ruleMatches
}

func (r *CommaWhitespaceRule) msg(key, def string) string {
	if r.Messages != nil {
		if s, ok := r.Messages[key]; ok && s != "" {
			return s
		}
	}
	return def
}

func (r *CommaWhitespaceRule) isDomain(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	return i < len(tokens) && r.domain.MatchString(tokens[i].GetToken())
}

func (r *CommaWhitespaceRule) isFileExtension(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	return i < len(tokens) && r.fileExt.MatchString(tokens[i].GetToken())
}

func isWhitespaceToken(token *languagetool.AnalyzedTokenReadings) bool {
	t := token.GetToken()
	return (token.IsWhitespace() || tools.IsNonBreakingWhitespace(t) || token.IsFieldCode()) && t != "\u200B"
}

func isQuote(str string) bool {
	if len([]rune(str)) != 1 {
		return false
	}
	c := []rune(str)[0]
	return c == '\'' || c == '"' || c == '’' || c == '”' || c == '“' || c == '«' || c == '»'
}

func isHyphenOrComma(str string) bool {
	if len([]rune(str)) != 1 {
		return false
	}
	c := []rune(str)[0]
	return c == '-' || c == ','
}

func isDigitOrDot(str string) bool {
	if str == "" {
		return false
	}
	c := []rune(str)[0]
	return c == '.' || unicode.IsDigit(c)
}

func isLeftBracket(str string) bool {
	if str == "" {
		return false
	}
	return []rune(str)[0] == '('
}

func isRightBracket(str string) bool {
	if str == "" {
		return false
	}
	return []rune(str)[0] == ')'
}

func containsDigit(str string) bool {
	for _, r := range str {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func utf16Substring(s string, from, to int) string {
	// Convert Java UTF-16 indices to Go string slice via runes for BMP-heavy tests
	// Full UTF-16: encode then slice
	u := []uint16{}
	for _, r := range s {
		if r >= 0x10000 {
			r -= 0x10000
			u = append(u, uint16(0xD800+(r>>10)), uint16(0xDC00+(r&0x3FF)))
		} else {
			u = append(u, uint16(r))
		}
	}
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	// decode back
	var runes []rune
	for i := from; i < to; {
		r := rune(u[i])
		if r >= 0xD800 && r <= 0xDBFF && i+1 < to {
			r2 := rune(u[i+1])
			if r2 >= 0xDC00 && r2 <= 0xDFFF {
				runes = append(runes, 0x10000+((r-0xD800)<<10)|(r2-0xDC00))
				i += 2
				continue
			}
		}
		runes = append(runes, r)
		i++
	}
	return string(runes)
}
