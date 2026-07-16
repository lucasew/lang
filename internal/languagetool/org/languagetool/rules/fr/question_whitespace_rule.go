package fr

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	espaceFineInsecable = "\u202F"
	nbsp                = "\u00A0"
)

var urlSchemeRE = regexp.MustCompile(`^(file|s?ftp|finger|git|gopher|hdl|https?|shttp|imap|mailto|mms|nntp|s?news(post|reply)?|prospero|rsync|rtspu|sips?|svn|svn\+ssh|telnet|wais)$`)

// QuestionWhitespaceRule ports org.languagetool.rules.fr.QuestionWhitespaceRule.
// Requires fine/nbsp spaces before ?!;: and around guillemets (non-strict: any
// whitespace before ?!; is accepted).
type QuestionWhitespaceRule struct {
	Messages map[string]string
	// Strict requires U+202F before ?!; (not implemented for twin; same as non-strict accept).
}

func NewQuestionWhitespaceRule(messages map[string]string) *QuestionWhitespaceRule {
	return &QuestionWhitespaceRule{Messages: messages}
}

func (r *QuestionWhitespaceRule) GetID() string { return "FRENCH_WHITESPACE" }

func (r *QuestionWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokens() // include whitespace tokens
	var ruleMatches []*rules.RuleMatch
	prevPrevToken := ""
	prevToken := ""
	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		if tokens[i].IsImmunized() || prevToken == "(" || prevToken == "[" {
			prevPrevToken = prevToken
			prevToken = token
			continue
		}
		// Light anti-patterns (smileys, times, ??, MAC, CSV)
		if frWhitespaceAntiPattern(tokens, i) {
			prevPrevToken = prevToken
			prevToken = token
			continue
		}

		var msg, suggestionText string
		iFrom, iTo := i-1, i
		isPreviousWhitespace := i > 0 && tokens[i-1].IsWhitespace()
		prevTokenToChange := prevToken
		if isPreviousWhitespace {
			prevTokenToChange = ""
		}
		if !isAllowedWhitespaceChar(tokens, i-1) {
			if token == "?" && prevToken != "!" {
				msg = "Le point d'interrogation est précédé d'une espace fine insécable."
				suggestionText = prevTokenToChange + espaceFineInsecable + "?"
			} else if token == "!" && prevToken != "?" {
				msg = "Le point d'exclamation est précédé d'une espace fine insécable."
				suggestionText = prevTokenToChange + espaceFineInsecable + "!"
			} else if token == ";" {
				msg = "Le point-virgule est précédé d'une espace fine insécable."
				suggestionText = prevTokenToChange + espaceFineInsecable + ";"
			} else if token == ":" {
				if !urlSchemeRE.MatchString(prevToken) {
					msg = "Les deux-points sont précédés d'une espace insécable."
					suggestionText = prevTokenToChange + nbsp + ":"
				}
			} else if token == "»" {
				if prevPrevToken == "«" {
					msg = "Les guillemets sont toujours accompagnés d'une espace insécable."
					suggestionText = "«" + nbsp + prevTokenToChange + nbsp + "»"
					iFrom = i - 2
				} else {
					msg = "Le guillemet fermant est précédé d'une espace insécable."
					suggestionText = prevTokenToChange + nbsp + "»"
				}
			}
		}

		if prevToken == "«" {
			if tools.IsEmptyStr(token) || token == "" {
				msg = "Le guillemet ouvrant est suivi d'une espace insécable."
				suggestionText = "«" + nbsp
				iTo = i - 1
			} else if !isAllowedWhitespaceChar(tokens, i) {
				nextToken := ""
				if i+1 < len(tokens) {
					nextToken = tokens[i+1].GetToken()
				}
				if nextToken != "»" {
					msg = "Le guillemet ouvrant est suivi d'une espace insécable."
					if !tokens[i].IsWhitespace() {
						suggestionText = "«" + nbsp + token
					} else {
						suggestionText = "«" + nbsp
					}
				}
			}
		}

		if msg != "" {
			fromPos := tokens[iFrom].GetStartPos()
			toPos := tokens[iTo].GetEndPos()
			rm := rules.NewRuleMatch(r, sentence, fromPos, toPos, msg)
			rm.ShortMessage = "Insérer une espace insécable"
			rm.SetSuggestedReplacement(suggestionText)
			ruleMatches = append(ruleMatches, rm)
		}
		prevPrevToken = prevToken
		prevToken = token
	}
	return ruleMatches
}

func isAllowedWhitespaceChar(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i < 0 || i >= len(tokens) {
		return false
	}
	return tokens[i].IsWhitespace()
}

// frWhitespaceAntiPattern approximates ANTI_PATTERNS without DisambiguationPatternRule.
func frWhitespaceAntiPattern(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	tok := tokens[i].GetToken()
	// smileys :-) :)
	if (tok == ")" || tok == "(" || tok == "D") && i >= 1 {
		if tokens[i-1].GetToken() == "-" && i >= 2 {
			p := tokens[i-2].GetToken()
			if p == ":" || p == ";" {
				return true
			}
		}
		if i >= 1 {
			p := tokens[i-1].GetToken()
			if (p == ":" || p == ";") && !tokens[i].IsWhitespaceBefore() {
				return true
			}
		}
	}
	// times 23:20
	if tok == ":" && i > 0 && i+1 < len(tokens) {
		prev, next := tokens[i-1].GetToken(), tokens[i+1].GetToken()
		if looksLikeTimePart(prev) && isDigits(next) {
			return true
		}
	}
	// ?? !!
	if (tok == "?" || tok == "!") && i > 0 {
		p := tokens[i-1].GetToken()
		if p == "?" || p == "!" {
			return true
		}
	}
	// MAC-ish : between hex pairs
	if tok == ":" && i > 0 && i+1 < len(tokens) {
		if isHex2(tokens[i-1].GetToken()) && isHex2(tokens[i+1].GetToken()) {
			return true
		}
	}
	// CSV ;
	if tok == ";" {
		if i+1 < len(tokens) && !tokens[i+1].IsWhitespaceBefore() && tokens[i+1].GetToken() != "" {
			// 1;2;3 style
			return true
		}
	}
	return false
}

func looksLikeTimePart(s string) bool {
	// ends with 1-2 digits
	n := 0
	for i := len(s) - 1; i >= 0 && n < 3; i-- {
		if s[i] >= '0' && s[i] <= '9' {
			n++
		} else {
			break
		}
	}
	return n >= 1 && n <= 2
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isHex2(s string) bool {
	if len(s) != 2 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// silence
var _ = strings.Contains
