package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleRepeatedSentenceBeginning is a surface stand-in for
// org.languagetool.rules.de.StyleRepeatedSentenceBeginning.
// Without ART/PRO POS tags, flags ≥3 consecutive sentences that start with a
// definite/indefinite article or a nominative personal pronoun.
type StyleRepeatedSentenceBeginning struct {
	Messages    map[string]string
	MinRepeated int
}

func NewStyleRepeatedSentenceBeginning(messages map[string]string) *StyleRepeatedSentenceBeginning {
	return &StyleRepeatedSentenceBeginning{Messages: messages, MinRepeated: 3}
}

func (r *StyleRepeatedSentenceBeginning) GetID() string {
	return "STYLE_REPEATED_SENTENCE_BEGINNING"
}

var styleBeginArticles = map[string]struct{}{
	"der": {}, "die": {}, "das": {}, "den": {}, "dem": {}, "des": {},
	"ein": {}, "eine": {}, "einen": {}, "einem": {}, "einer": {}, "eines": {},
}

var styleBeginPronouns = map[string]struct{}{
	"ich": {}, "du": {}, "er": {}, "sie": {}, "es": {}, "wir": {}, "ihr": {},
	"man": {},
}

func styleIsSubjectStart(token string) (is bool, endAtWord bool) {
	lc := strings.ToLower(token)
	if _, ok := styleBeginArticles[lc]; ok {
		return true, true // extend to following noun-ish word
	}
	if _, ok := styleBeginPronouns[lc]; ok {
		return true, false
	}
	return false, false
}

func (r *StyleRepeatedSentenceBeginning) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
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
			msg := "Subjekt als wiederholter Satzanfang"
			for i := 0; i < len(repeated); i++ {
				rm := rules.NewRuleMatch(r, repeated[i], startPos[i], endPos[i], msg)
				rm.ShortMessage = "wiederholter Satzanfang"
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
		if len(tokens) < 2 {
			flush()
			pos += sentence.GetCorrectedTextLength()
			continue
		}
		first := tokens[1]
		ok, extend := styleIsSubjectStart(first.GetToken())
		if !ok {
			flush()
			pos += sentence.GetCorrectedTextLength()
			continue
		}
		from := first.GetStartPos() + pos
		to := first.GetEndPos() + pos
		if extend {
			// include next content token if present (article + noun)
			if len(tokens) > 2 && !strings.ContainsAny(tokens[2].GetToken(), ".?!") {
				to = tokens[2].GetEndPos() + pos
			}
		}
		repeated = append(repeated, sentence)
		startPos = append(startPos, from)
		endPos = append(endPos, to)
		nRepeated++
		pos += sentence.GetCorrectedTextLength()
	}
	flush()
	return ruleMatches
}
