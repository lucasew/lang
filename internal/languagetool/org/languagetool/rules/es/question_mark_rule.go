package es

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// QuestionMarkRule ports org.languagetool.rules.es.QuestionMarkRule.
// POS-based repositioning after commas uses surface heuristics (no tagger).
type QuestionMarkRule struct {
	Messages map[string]string
}

func NewQuestionMarkRule(messages map[string]string) *QuestionMarkRule {
	return &QuestionMarkRule{Messages: messages}
}

func (r *QuestionMarkRule) GetID() string { return "ES_QUESTION_MARK" }

// MatchList is the text-level entry point.
func (r *QuestionMarkRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var matches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		tokens := sentence.GetTokensWithoutWhitespace()
		needsQ := hasTokenAtPos("?", tokens)
		needsE := hasTokenAtPos("!", tokens)
		if needsQ > 1 || needsE > 1 {
			hasInvQ, hasInvE := false, false
			var firstToken *languagetool.AnalyzedTokenReadings
			for i := 0; i < len(tokens); i++ {
				tok := tokens[i].GetToken()
				if firstToken == nil && !tokens[i].IsSentenceStart() && isESContentWord(tok) {
					firstToken = tokens[i]
				}
				if tok == "¿" && i < needsQ {
					hasInvQ = true
				} else if tok == "¡" && i < needsE {
					hasInvE = true
				}
				if !tokens[i].IsSentenceEnd() &&
					((tok == "?" && i > needsQ) || (tok == "!" && i > needsE)) {
					firstToken = nil
				}
				// After ":" (e.g. "Marco: Puedes…") start question at following word.
				if i > 0 && tokens[i-1].GetToken() == ":" && isESContentWord(tok) {
					firstToken = tokens[i]
				}
				// After comma: start inverted mark at question fragment (surface heuristics).
				if i > 0 && tokens[i-1].GetToken() == "," {
					tl := strings.ToLower(tok)
					if isESQuestionStart(tl) || tl == "no" || tl == "sí" || tl == "eh" || tl == "de" {
						firstToken = tokens[i]
					}
					if (tl == "pero" || tl == "o") && i+1 < len(tokens) {
						n := strings.ToLower(tokens[i+1].GetToken())
						if isESQuestionStart(n) {
							firstToken = tokens[i+1]
						} else if n == "no" || n == "sí" || n == "eh" {
							firstToken = tokens[i+1]
						} else if tl == "o" && (n == "no" || n == "sí") {
							firstToken = tokens[i]
						}
					} else if tl == "pero" {
						// "Pero cómo" without treating cómo specially if next not question
						// "Pero, cómo" handled above with next
					}
					// "o no?" → firstToken o already if only o; test wants ¿o
					if tl == "o" || tl == "no" || tl == "eh" {
						firstToken = tokens[i]
					}
				}
			}
			if firstToken != nil {
				var s string
				if needsQ > 1 && needsE > 1 {
					// skip ¡¿...?!
				} else if needsQ > 1 && !hasInvQ {
					s = "¿"
				} else if needsE > 1 && !hasInvE {
					s = "¡"
				}
				if s != "" {
					msg := "Símbolo desparejado: Parece que falta un '" + s + "'"
					rm := rules.NewRuleMatch(r, sentence, pos+firstToken.GetStartPos(), pos+firstToken.GetEndPos(), msg)
					rm.SetSuggestedReplacement(s + firstToken.GetToken())
					matches = append(matches, rm)
				}
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return matches
}

func isESContentWord(tok string) bool {
	if tok == "" || tools.IsPunctuationMark(tok) {
		return false
	}
	// skip dashes/quotes used as dialogue markers
	switch tok {
	case "—", "–", "-", "\"", "«", "»", "'", "“", "”":
		return false
	}
	for _, r := range tok {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func isESQuestionStart(tl string) bool {
	switch tl {
	case "qué", "que", "cómo", "como", "cuál", "cual", "cuáles", "quién", "quien",
		"dónde", "donde", "cuándo", "cuando", "cuánto", "cuanto", "por":
		return true
	}
	return false
}

func hasTokenAtPos(ch string, tokens []*languagetool.AnalyzedTokenReadings) int {
	for i := len(tokens) - 1; i > 0; i-- {
		if tokens[i].GetToken() == ch {
			if i < len(tokens)-1 && !tokens[i+1].IsWhitespaceBefore() &&
				!tools.IsPunctuationMark(tokens[i+1].GetToken()) &&
				!tokens[i+1].IsWhitespace() {
				continue
			}
			return i
		}
	}
	return -1
}
