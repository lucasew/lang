package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CatalanUnpairedQuestionMarksRule ports org.languagetool.rules.ca.CatalanUnpairedQuestionMarksRule.
// Comma-clause repositioning is POS-gated (Java FreeLing tags); without tags those arms
// fail closed (no surface invent of question-word lists). Default-off in Java.
// CatalanUnpairedExclamationMarksRule is the ¡/! twin via NewCatalanUnpairedExclamationMarksRule.
type CatalanUnpairedQuestionMarksRule struct {
	Messages map[string]string
	start    string
	end      string
	id       string
	desc     string
	// issueType: Java CatalanUnpairedQuestionMarksRule uses Typographical default;
	// CatalanUnpairedExclamationMarksRule sets Style.
	issueType rules.ITSIssueType
	// minToCheckParagraph: Java Exclamation sets 1; Question uses base default.
	minToCheckParagraph int
}

func NewCatalanUnpairedQuestionMarksRule(messages map[string]string) *CatalanUnpairedQuestionMarksRule {
	return &CatalanUnpairedQuestionMarksRule{
		Messages:  messages,
		start:     "¿",
		end:       "?",
		id:        "CA_UNPAIRED_QUESTION",
		desc:      "Exigeix signe d'interrogació inicial",
		issueType: rules.ITSTypographical,
	}
}

// NewCatalanUnpairedExclamationMarksRule ports CatalanUnpairedExclamationMarksRule.
func NewCatalanUnpairedExclamationMarksRule(messages map[string]string) *CatalanUnpairedQuestionMarksRule {
	return &CatalanUnpairedQuestionMarksRule{
		Messages:            messages,
		start:               "¡",
		end:                 "!",
		id:                  "CA_UNPAIRED_EXCLAMATION",
		desc:                "Exigeix signe d'exclamació inicial",
		issueType:           rules.ITSStyle,
		minToCheckParagraph: 1,
	}
}

func (r *CatalanUnpairedQuestionMarksRule) GetID() string          { return r.id }
func (r *CatalanUnpairedQuestionMarksRule) GetDescription() string { return r.desc }
func (r *CatalanUnpairedQuestionMarksRule) IsDefaultOff() bool     { return true }

func (r *CatalanUnpairedQuestionMarksRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.issueType == "" {
		return rules.ITSTypographical
	}
	return r.issueType
}

func (r *CatalanUnpairedQuestionMarksRule) MinToCheckParagraph() int {
	if r == nil {
		return 0
	}
	return r.minToCheckParagraph
}

// MatchList ports CatalanUnpairedQuestionMarksRule.match.
func (r *CatalanUnpairedQuestionMarksRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var matches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		tokens := sentence.GetTokensWithoutWhitespace()
		if len(tokens) == 0 {
			continue
		}
		needsAt := hasEndSymbolAt(r.end, tokens)
		if needsAt > 1 {
			hasStart := false
			var firstToken *languagetool.AnalyzedTokenReadings
			for i := 0; i < len(tokens); i++ {
				tok := tokens[i].GetToken()
				// Java: first non-SENT_START, non-punctuation
				if firstToken == nil && !tokens[i].IsSentenceStart() && !tools.IsPunctuationMark(tok) {
					firstToken = tokens[i]
				}
				if tok == r.start && i < needsAt {
					hasStart = true
				}
				// possibly a sentence end (Java: end symbol with i < needsAt)
				if !tokens[i].IsSentenceEnd() && tok == r.end && i < needsAt {
					firstToken = nil
				}
				// put the question mark in: ¿de què… (FreeLing POS)
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
				// Java surface: no | sí | oi | eh after comma
				if i > 2 && i < len(tokens) {
					if tokens[i-1].GetToken() == "," &&
						(tokens[i].GetToken() == "no" || tokens[i].GetToken() == "sí" ||
							tokens[i].GetToken() == "oi" || tokens[i].GetToken() == "eh") {
						firstToken = tokens[i]
					}
				}
			}
			if firstToken != nil && !hasStart {
				s := r.start
				msg := "Símbol sense parella: Sembla que falta un '" + s + "'"
				rm := rules.NewRuleMatch(r, sentence, pos+firstToken.GetStartPos(), pos+firstToken.GetEndPos(), msg)
				rm.SetSuggestedReplacement(s + firstToken.GetToken())
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
			!tools.IsPunctuationMark(tokens[i+1].GetToken()) && !tokens[i+1].IsWhitespace() {
			continue // URL-like glued mark
		}
		return i
	}
	return -1
}
