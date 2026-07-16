package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleRepeatedVeryShortSentences ports org.languagetool.rules.de.StyleRepeatedVeryShortSentences
// without direct-speech exclusion (surface length only).
type StyleRepeatedVeryShortSentences struct {
	Messages    map[string]string
	MinWords    int
	MinRepeated int
}

func NewStyleRepeatedVeryShortSentences(messages map[string]string) *StyleRepeatedVeryShortSentences {
	return &StyleRepeatedVeryShortSentences{
		Messages:    messages,
		MinWords:    4,
		MinRepeated: 3,
	}
}

func (r *StyleRepeatedVeryShortSentences) GetID() string { return "STYLE_REPEATED_SHORT_SENTENCES" }

func (r *StyleRepeatedVeryShortSentences) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	minW := r.MinWords
	if minW <= 0 {
		minW = 4
	}
	minR := r.MinRepeated
	if minR <= 0 {
		minR = 3
	}
	if len(sentences) < minR {
		return nil
	}
	var ruleMatches []*rules.RuleMatch
	pos := 0
	nRepeated := 0
	var startPos, endPos []int
	var repeated []*languagetool.AnalyzedSentence
	flush := func() {
		if nRepeated >= minR {
			for i := 0; i < len(repeated); i++ {
				msg := "Stakkato-Sätze"
				rm := rules.NewRuleMatch(r, repeated[i], startPos[i], endPos[i], msg)
				rm.ShortMessage = "kurze Sätze"
				ruleMatches = append(ruleMatches, rm)
			}
		}
		repeated = nil
		startPos = nil
		endPos = nil
		nRepeated = 0
	}
	for _, sentence := range sentences {
		tokens := sentence.GetTokensWithoutWhitespace()
		// tokens include SENT_START; short sentence: >3 and <= minWords+2 (start + words + punct)
		if len(tokens) > 3 && len(tokens) <= minW+2 {
			repeated = append(repeated, sentence)
			// mark last content token through end punctuation (Java: tokens[len-2] .. tokens[len-1])
			from := tokens[len(tokens)-2].GetStartPos() + pos
			to := tokens[len(tokens)-1].GetEndPos() + pos
			startPos = append(startPos, from)
			endPos = append(endPos, to)
			nRepeated++
		} else {
			flush()
		}
		pos += sentence.GetCorrectedTextLength()
	}
	flush()
	return ruleMatches
}
