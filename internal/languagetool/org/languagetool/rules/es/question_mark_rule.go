package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// QuestionMarkRule ports org.languagetool.rules.es.QuestionMarkRule.
// Comma-clause repositioning is POS-gated (Java FreeLing tags); without tags those
// arms fail closed (no surface invent of question-word lists).
type QuestionMarkRule struct {
	Messages map[string]string
}

func NewQuestionMarkRule(messages map[string]string) *QuestionMarkRule {
	return &QuestionMarkRule{Messages: messages}
}

func (r *QuestionMarkRule) GetID() string { return "ES_QUESTION_MARK" }

// MatchList ports QuestionMarkRule.match(List<AnalyzedSentence>).
func (r *QuestionMarkRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var matches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		tokens := sentence.GetTokensWithoutWhitespace()
		if len(tokens) == 0 {
			continue
		}
		needsQ := hasTokenAtPos("?", tokens)
		needsE := hasTokenAtPos("!", tokens)
		if needsQ > 1 || needsE > 1 {
			hasInvQ, hasInvE := false, false
			var firstToken *languagetool.AnalyzedTokenReadings
			for i := 0; i < len(tokens); i++ {
				tok := tokens[i].GetToken()
				// Java: first non-SENT_START, non-punctuation
				if firstToken == nil && !tokens[i].IsSentenceStart() && !tools.IsPunctuationMark(tok) {
					firstToken = tokens[i]
				}
				if tok == "¿" && i < needsQ {
					hasInvQ = true
				} else if tok == "¡" && i < needsE {
					hasInvE = true
				}
				// possibly a sentence end (extra ?/! later in the sentence)
				if !tokens[i].IsSentenceEnd() &&
					((tok == "?" && i > needsQ) || (tok == "!" && i > needsE)) {
					firstToken = nil
				}
				// Colon often marks a clause boundary (Java SRX may split; without it reset firstToken).
				if tok == ":" {
					firstToken = nil
				}
				// put the question mark in: ¿de qué… ¿para cuál… ¿cómo…
				// Java FreeLing POS: CC, SPS00, PT*, DT*
				if i > 2 && i+2 < len(tokens) {
					if tokens[i-1].GetToken() == "," && tokens[i].HasPosTag("CC") && tokens[i+1].HasPosTag("SPS00") &&
						(tokens[i+2].HasPosTagStartingWith("PT") || tokens[i+2].HasPosTagStartingWith("DT")) {
						firstToken = tokens[i]
					}
					if tokens[i-1].GetToken() == "," && tokens[i].HasPosTag("SPS00") &&
						(tokens[i+1].HasPosTagStartingWith("PT") || tokens[i+1].HasPosTagStartingWith("DT")) {
						firstToken = tokens[i]
					}
					if tokens[i-1].GetToken() == "," && tokens[i].HasPosTag("CC") &&
						(tokens[i+1].HasPosTagStartingWith("PT") || tokens[i+1].HasPosTagStartingWith("DT")) {
						firstToken = tokens[i]
					}
					if tokens[i-1].GetToken() == "," &&
						(tokens[i].HasPosTagStartingWith("PT") || tokens[i].HasPosTagStartingWith("DT")) {
						firstToken = tokens[i]
					}
					if tokens[i-1].GetToken() == "," && tokens[i].HasPosTag("CC") &&
						(tokens[i+1].GetToken() == "no" || tokens[i+1].GetToken() == "sí") {
						firstToken = tokens[i]
					}
				}
				if i > 2 && i < len(tokens) {
					if tokens[i-1].GetToken() == "," &&
						(tokens[i].GetToken() == "no" || tokens[i].GetToken() == "sí" || tokens[i].GetToken() == "eh") {
						firstToken = tokens[i]
					}
				}
			}
			if firstToken != nil && !firstToken.HasPosTag("_english_ignore_") {
				var s string
				if needsQ > 1 && needsE > 1 {
					// ignore for now, e.g. "¡¿Nunca tienes clases o qué?!"
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

func hasTokenAtPos(ch string, tokens []*languagetool.AnalyzedTokenReadings) int {
	for i := len(tokens) - 1; i > 0; i-- {
		if tokens[i].GetToken() == ch {
			if i < len(tokens)-1 && !tokens[i+1].IsWhitespaceBefore() &&
				!tools.IsPunctuationMark(tokens[i+1].GetToken()) &&
				!tokens[i+1].IsWhitespace() {
				// ignore question marks joined to the next word (URLs)
				continue
			}
			return i
		}
	}
	return -1
}
