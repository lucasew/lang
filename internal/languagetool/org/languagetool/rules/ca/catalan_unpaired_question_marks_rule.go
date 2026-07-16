package ca

import (
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CatalanUnpairedQuestionMarksRule ports org.languagetool.rules.ca.CatalanUnpairedQuestionMarksRule
// without POS-based re-anchoring after commas.
type CatalanUnpairedQuestionMarksRule struct {
	Messages map[string]string
	start    string
	end      string
	id       string
	desc     string
}

func NewCatalanUnpairedQuestionMarksRule(messages map[string]string) *CatalanUnpairedQuestionMarksRule {
	return &CatalanUnpairedQuestionMarksRule{
		Messages: messages,
		start:    "¿",
		end:      "?",
		id:       "CA_UNPAIRED_QUESTION",
		desc:     "Exigeix signe d'interrogació inicial",
	}
}

func NewCatalanUnpairedExclamationMarksRule(messages map[string]string) *CatalanUnpairedQuestionMarksRule {
	return &CatalanUnpairedQuestionMarksRule{
		Messages: messages,
		start:    "¡",
		end:      "!",
		id:       "CA_UNPAIRED_EXCLAMATION",
		desc:     "Exigeix signe d'exclamació inicial",
	}
}

func (r *CatalanUnpairedQuestionMarksRule) GetID() string { return r.id }

func (r *CatalanUnpairedQuestionMarksRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var matches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		tokens := sentence.GetTokensWithoutWhitespace()
		needsAt := hasEndSymbolAt(r.end, tokens)
		if needsAt > 1 {
			hasStart := false
			var first *languagetool.AnalyzedTokenReadings
			for i := 0; i < len(tokens); i++ {
				tok := tokens[i].GetToken()
				if first == nil && !tokens[i].IsSentenceStart() && !isPunctMark(tok) {
					first = tokens[i]
				}
				if tok == r.start && i < needsAt {
					hasStart = true
				}
			}
			if first != nil && !hasStart {
				msg := "Símbol sense parella: Sembla que falta un '" + r.start + "'"
				rm := rules.NewRuleMatch(r, sentence, pos+first.GetStartPos(), pos+first.GetEndPos(), msg)
				rm.SetSuggestedReplacement(r.start + first.GetToken())
				matches = append(matches, rm)
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return matches
}

func hasEndSymbolAt(ch string, tokens []*languagetool.AnalyzedTokenReadings) int {
	for i := len(tokens) - 1; i > 0; i-- {
		if tokens[i].GetToken() != ch {
			continue
		}
		if i < len(tokens)-1 && !tokens[i+1].IsWhitespaceBefore() &&
			!isPunctMark(tokens[i+1].GetToken()) && !tokens[i+1].IsWhitespace() {
			continue // URL-like glued mark
		}
		return i
	}
	return -1
}

func isPunctMark(s string) bool {
	if s == "" {
		return false
	}
	// tools may not export IsPunctuationMark; approximate
	if tools.IsAllUppercase(s) && len([]rune(s)) == 1 {
		// not punct
	}
	for _, r := range s {
		if !unicode.IsPunct(r) && r != '¿' && r != '¡' {
			return false
		}
	}
	return true
}
