package km

import (
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// KhmerWordRepeatRule ports org.languagetool.rules.km.KhmerWordRepeatRule.
// Ignores repeats separated by a normal space (U+0020); ZWSP-separated repeats flag.
type KhmerWordRepeatRule struct {
	Messages map[string]string
}

func NewKhmerWordRepeatRule(messages map[string]string) *KhmerWordRepeatRule {
	return &KhmerWordRepeatRule{Messages: messages}
}

func (r *KhmerWordRepeatRule) GetID() string { return "KM_WORD_REPEAT_RULE" }

func (r *KhmerWordRepeatRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	tokensWithWS := sentence.GetTokens()
	// map non-blank index → full-token index
	origPos := mapNonBlankToFull(tokensWithWS)

	prevToken := ""
	msg := r.Messages["repetition"]
	if msg == "" {
		msg = "Word repetition"
	}
	short := r.Messages["desc_repetition_short"]
	if short == "" {
		short = "Repetition"
	}

	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		if isKhmerWord(token) && equalFold(prevToken, token) && !r.ignore(tokensWithWS, origPos, i) {
			prevPos := tokens[i-1].GetStartPos()
			pos := tokens[i].GetStartPos()
			rm := rules.NewRuleMatch(r, sentence, prevPos, pos+utf16LenKM(prevToken), msg)
			rm.ShortMessage = short
			rm.SetSuggestedReplacements([]string{
				prevToken + " " + token,
				prevToken,
				prevToken + "ៗ",
			})
			ruleMatches = append(ruleMatches, rm)
		}
		prevToken = token
	}
	return ruleMatches
}

func mapNonBlankToFull(tokensWithWS []*languagetool.AnalyzedTokenReadings) []int {
	var m []int
	for i, t := range tokensWithWS {
		if !t.IsWhitespace() || t.IsSentenceStart() || t.IsSentenceEnd() || t.IsParagraphEnd() {
			m = append(m, i)
		}
	}
	return m
}

func (r *KhmerWordRepeatRule) ignore(tokensWithWS []*languagetool.AnalyzedTokenReadings, origPos []int, position int) bool {
	if position < 1 || position >= len(origPos) {
		return false
	}
	fullIdx := origPos[position]
	if fullIdx >= 1 && tokensWithWS[fullIdx-1].GetToken() == "\u0020" {
		return true
	}
	return false
}

func isKhmerWord(token string) bool {
	runes := []rune(token)
	if len(runes) == 1 && !unicode.IsLetter(runes[0]) {
		return false
	}
	return true
}

func equalFold(a, b string) bool {
	return a == b || (len(a) == len(b) && a == b) // Khmer case-less; keep exact
}

func utf16LenKM(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
