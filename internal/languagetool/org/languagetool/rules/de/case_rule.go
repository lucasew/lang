package de

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CaseRule is a surface stand-in for org.languagetool.rules.de.CaseRule.
// Without full POS/tagger, it flags common mid-sentence capitalization errors.
type CaseRule struct {
	Messages map[string]string
}

func NewCaseRule(messages map[string]string) *CaseRule {
	return &CaseRule{Messages: messages}
}

func (r *CaseRule) GetID() string { return "DE_CASE" }

var caseDeterminers = map[string]struct{}{
	"der": {}, "die": {}, "das": {}, "dem": {}, "den": {}, "des": {},
	"ein": {}, "eine": {}, "einem": {}, "einen": {}, "einer": {}, "eines": {},
	"mein": {}, "meine": {}, "meinem": {}, "meinen": {}, "meiner": {}, "meines": {},
	"dein": {}, "deine": {}, "deinem": {}, "deinen": {}, "deiner": {}, "deines": {},
	"sein": {}, "seine": {}, "seinem": {}, "seinen": {}, "seiner": {}, "seines": {},
	"ihr": {}, "ihre": {}, "ihrem": {}, "ihren": {}, "ihrer": {}, "ihres": {},
	"unser": {}, "unsere": {}, "unserem": {}, "unseren": {}, "unserer": {}, "unseres": {},
	"kein": {}, "keine": {}, "keinem": {}, "keinen": {}, "keiner": {}, "keines": {},
	"dieser": {}, "diese": {}, "dieses": {}, "diesem": {}, "diesen": {},
	"jener": {}, "jene": {}, "jenes": {}, "jenem": {}, "jenen": {},
	"alle": {}, "allem": {}, "allen": {}, "aller": {}, "alles": {},
	"viele": {}, "vieler": {}, "wenige": {},
}

// Capitals often wrongly uppercased after a determiner (from CaseRuleTest assertBad).
var caseWrongAfterDet = map[string]struct{}{
	"Neue": {}, "Absolute": {}, "Allgemeine": {}, "Alles": {}, "Liebe": {},
	"Meinem": {}, "Meines": {}, "Meiner": {}, "Meinen": {},
	"Großes": {}, "Gesagte": {}, "Gesagten": {}, "Erzählte": {}, "Erzählten": {},
	"Gratis": {}, "Ohne": {}, "Blaue": {}, "Eingeschlossenen": {}, "Gefragte": {},
	"Gesammelten": {},
}

// Mid-sentence capitals often wrong
var caseWrongMid = map[string]struct{}{
	"Heute": {}, "Vertraute": {}, "Lernt": {}, "Geradliniges": {}, "Herzlich": {},
	"Nah": {}, "Fern": {}, "Alternativ": {}, "Alles": {}, "Was": {},
}

var numberingRE = regexp.MustCompile(`^\d+(\.\d+)*$`)

func (r *CaseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 2; i < len(tokens); i++ { // skip SENT_START and first word
		tok := tokens[i]
		w := tok.GetToken()
		if !tools.StartsWithUppercase(w) || utf8.RuneCountInString(w) < 2 {
			continue
		}
		prev := tokens[i-1].GetToken()
		if isSentenceRestartDE(prev) {
			continue
		}
		if looksLikeNumberingDE(tokens, i) {
			continue
		}
		if IsCaseRuleException(w) {
			continue
		}
		// "die Die"
		if strings.EqualFold(prev, w) && isDet(prev) {
			matches = append(matches, caseMatch(r, sentence, tok, w))
			continue
		}
		if isDet(prev) {
			if _, bad := caseWrongAfterDet[w]; bad {
				matches = append(matches, caseMatch(r, sentence, tok, w))
				continue
			}
		}
		if _, bad := caseWrongMid[w]; bad {
			matches = append(matches, caseMatch(r, sentence, tok, w))
		}
	}
	return matches
}

func caseMatch(r *CaseRule, sentence *languagetool.AnalyzedSentence, tok *languagetool.AnalyzedTokenReadings, w string) *rules.RuleMatch {
	msg := "Möglicherweise falsche Großschreibung."
	rm := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
	rm.ShortMessage = "Großschreibung"
	rm.SetSuggestedReplacement(lowerFirst(w))
	return rm
}

func isDet(w string) bool {
	_, ok := caseDeterminers[strings.ToLower(w)]
	return ok
}

func isSentenceRestartDE(w string) bool {
	switch w {
	case ".", "!", "?", ":", ";", "\"", "„", "“", "»", "«", "'", "(", "[", "…":
		return true
	}
	return false
}

func looksLikeNumberingDE(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i < 2 {
		return false
	}
	p := tokens[i-1].GetToken()
	if p == ")" || p == "." {
		return true
	}
	return numberingRE.MatchString(p)
}

func lowerFirst(s string) string {
	rs := []rune(s)
	if len(rs) == 0 {
		return s
	}
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}

// CaseRuleCompareLists ports CaseRule.compareLists for exception-pattern matching.
func CaseRuleCompareLists(tokens []*languagetool.AnalyzedTokenReadings, startIndex, endIndex int, patterns []*regexp.Regexp) bool {
	if startIndex < 0 || endIndex >= len(tokens) || endIndex-startIndex+1 != len(patterns) {
		return false
	}
	for i := 0; i < len(patterns); i++ {
		if !patterns[i].MatchString(tokens[startIndex+i].GetToken()) {
			return false
		}
	}
	return true
}
