package de

import (
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// commonGermanWords ports HunspellRule.commonGermanWords (skip wrong-split).
var commonGermanWords = map[string]struct{}{
	"-": {}, "das": {}, "sein": {}, "mein": {}, "meine": {}, "meinen": {}, "meines": {}, "meiner": {},
	"haben": {}, "kein": {}, "keine": {}, "keinen": {}, "keinem": {}, "keines": {}, "keiner": {},
	"ein": {}, "eines": {}, "eins": {}, "einen": {}, "einem": {}, "eine": {}, "einer": {},
	"rund": {}, "sehr": {}, "mach": {}, "noch": {}, "nein": {}, "ja": {}, "hallo": {}, "hi": {},
	"die": {}, "der": {}, "den": {}, "dem": {}, "des": {}, "nacht": {},
	"diesen": {}, "dieser": {}, "dies": {}, "dieses": {}, "diesem": {},
	"zum": {}, "zur": {}, "beim": {}, "nichts": {},
	"aufs": {}, "aufm": {}, "aufn": {}, "ausn": {}, "ausm": {}, "aus": {},
	"fürs": {}, "für": {}, "osten": {}, "rein": {}, "raus": {}, "namen": {}, "shippen": {},
	"amt": {}, "wir": {},
}

// reLowerCaseWord ports GermanSpellerRule.LOWER_CASE_WORD: [a-zöäü]-.+
var reLowerCaseWord = regexp.MustCompile(`^[a-zöäü]-.+`)

// tryWrongSplitSuggestions ports HunspellRule wrong-split for DE Match.
func (r *GermanSpellerRule) tryWrongSplitSuggestions(
	sentence *languagetool.AnalyzedSentence,
	prevWord string,
	prevFrom int,
	word string,
	wordFrom int,
	cleanWord string,
) *rules.RuleMatch {
	if r == nil || prevWord == "" || word == "" {
		return nil
	}
	if _, ok := commonGermanWords[strings.ToLower(prevWord)]; ok {
		return nil
	}
	if _, ok := commonGermanWords[strings.ToLower(word)]; ok {
		return nil
	}

	// "thanky ou" -> "thank you"
	// Java: prevWord.substring(0, prevWord.length()-1) + prevWord.substring(length-1)+word
	// length/substring/charAt are UTF-16 code units.
	if pu := utf16EncodeDE(prevWord); len(pu) >= 1 {
		sugg1a := utf16DecodeDE(pu[:len(pu)-1])
		sugg1b := cutOffDot(utf16DecodeDE(pu[len(pu)-1:]) + word)
		if sugg1a != "" && sugg1b != "" &&
			!r.IsMisspelled(sugg1a) && !r.IsMisspelled(sugg1b) &&
			r.AcceptSuggestion(sugg1a+" "+sugg1b) {
			if m := r.createWrongSplitMatch(sentence, wordFrom, cleanWord, sugg1a, sugg1b, prevFrom); m != nil {
				return m
			}
		}
	}

	// "than kyou" -> "thank you"
	// Java: prevWord + word.charAt(0); word.substring(1)
	if wu := utf16EncodeDE(word); len(wu) > 1 {
		sugg2a := prevWord + utf16DecodeDE(wu[:1])
		sugg2b := cutOffDot(utf16DecodeDE(wu[1:]))
		if sugg2a != "" && sugg2b != "" &&
			!r.IsMisspelled(sugg2a) && !r.IsMisspelled(sugg2b) &&
			r.AcceptSuggestion(sugg2a+" "+sugg2b) {
			if m := r.createWrongSplitMatch(sentence, wordFrom, cleanWord, sugg2a, sugg2b, prevFrom); m != nil {
				return m
			}
		}
	}
	return nil
}

// createWrongSplitMatch ports SpellingCheckRule + GermanSpellerRule filter.
func (r *GermanSpellerRule) createWrongSplitMatch(
	sentence *languagetool.AnalyzedSentence,
	pos int,
	coveredWord, suggestion1, suggestion2 string,
	prevPos int,
) *rules.RuleMatch {
	if reLowerCaseWord.MatchString(suggestion2) {
		return nil
	}
	joined := strings.TrimSpace(suggestion1 + " " + suggestion2)
	msg := r.GetMessage(coveredWord, joined)
	// span prevPos .. pos+len(coveredWord) in UTF-16 units like Java
	m := rules.NewRuleMatch(r, sentence, prevPos, pos+utf16LenDE(coveredWord), msg)
	m.SetType(rules.RuleMatchTypeUnknownWord)
	m.SetShortMessage(r.spellingShortMessage())
	m.SetSuggestedReplacements([]string{joined})
	return m
}

// removeLastWrongSplitIfSamePrev drops the previous match when merging a wrong-split
// that covers the same prevPos (Java ruleMatchesSoFar.remove last if fromPos==prevPos).
func removeLastWrongSplitIfSamePrev(out []*rules.RuleMatch, prevPos int) []*rules.RuleMatch {
	if len(out) == 0 {
		return out
	}
	last := out[len(out)-1]
	if last != nil && last.GetFromPos() == prevPos {
		return out[:len(out)-1]
	}
	return out
}

// isCommonGermanWord reports membership in commonGermanWords (for tests).
func isCommonGermanWord(w string) bool {
	_, ok := commonGermanWords[strings.ToLower(w)]
	return ok
}

func utf16EncodeDE(s string) []uint16 {
	return utf16.Encode([]rune(s))
}

func utf16DecodeDE(u []uint16) string {
	if len(u) == 0 {
		return ""
	}
	return string(utf16.Decode(u))
}
