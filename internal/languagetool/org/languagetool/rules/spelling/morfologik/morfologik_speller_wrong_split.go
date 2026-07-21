package morfologik

import (
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// maxFrequencyForSplitting ports MorfologikSpellerRule.MAX_FREQUENCY_FOR_SPLITTING (0..21).
const maxFrequencyForSplitting = 21

// digitRunes ports StringUtils.containsAny(..., "0".."9") for wrong-split skip.
const digitRunes = "0123456789"

// getSpellerFrequency ports getFrequency(speller1, word).
func (r *MorfologikSpellerRule) getSpellerFrequency(word string) int {
	if r == nil {
		return 0
	}
	if r.Multi != nil {
		return r.Multi.GetFrequency(word)
	}
	if r.Speller == nil {
		return 0
	}
	return r.Speller.GetFrequency(word)
}

// createWrongSplitMatch ports SpellingCheckRule.createWrongSplitMatch via shared twin.
func (r *MorfologikSpellerRule) createWrongSplitMatch(
	sentence *languagetool.AnalyzedSentence,
	ruleMatches *[]*rules.RuleMatch,
	pos int,
	coveredWord, suggestion1, suggestion2 string,
	prevPos int,
) *rules.RuleMatch {
	return spelling.CreateWrongSplitMatch(r, sentence, ruleMatches, pos, coveredWord, suggestion1, suggestion2, prevPos)
}

// tryWrongSplitPrev ports getRuleMatches wrong-split with previous token.
// Returns match, beforeSuggestionStr ("prevWord "), and earlyReturn when prev is misspelled.
func (r *MorfologikSpellerRule) tryWrongSplitPrev(
	sentence *languagetool.AnalyzedSentence,
	ruleMatches *[]*rules.RuleMatch,
	idx int,
	tokens []*languagetool.AnalyzedTokenReadings,
	word string,
	startPos int,
) (match *rules.RuleMatch, beforeStr string, early bool) {
	if r == nil || idx <= 0 || word == "" || tokens == nil {
		return nil, "", false
	}
	tok := tokens[idx]
	if tok == nil || !tok.IsWhitespaceBefore() {
		return nil, "", false
	}
	prevTok := tokens[idx-1]
	if prevTok == nil {
		return nil, "", false
	}
	prevWord := prevTok.GetToken()
	if prevWord == "" || strings.ContainsAny(prevWord, digitRunes) {
		return nil, "", false
	}
	if r.getSpellerFrequency(prevWord) >= maxFrequencyForSplitting {
		return nil, "", false
	}
	prevStartPos := prevTok.GetStartPos()
	var ruleMatch *rules.RuleMatch

	// "thanky ou" -> "thank you"
	if pu := utf16.Encode([]rune(prevWord)); len(pu) >= 1 {
		sugg1a := string(utf16.Decode(pu[:len(pu)-1]))
		sugg1b := string(utf16.Decode(pu[len(pu)-1:])) + word
		if utf16LenMF(sugg1a) > 1 && utf16LenMF(sugg1b) > 2 &&
			!r.isMisspelledWord(sugg1a) && !r.isMisspelledWord(sugg1b) &&
			r.getSpellerFrequency(sugg1a)+r.getSpellerFrequency(sugg1b) > r.getSpellerFrequency(prevWord) {
			ruleMatch = r.createWrongSplitMatch(sentence, ruleMatches, startPos, word, sugg1a, sugg1b, prevStartPos)
			beforeStr = prevWord + " "
		}
	}

	// "than kyou" -> "thank you" ; but not "She awaked" -> "Shea waked"
	if wu := utf16.Encode([]rune(word)); len(wu) > 1 {
		sugg2a := prevWord + string(utf16.Decode(wu[:1]))
		sugg2b := string(utf16.Decode(wu[1:]))
		if utf16LenMF(sugg2b) > 2 && !r.isMisspelledWord(sugg2a) && !r.isMisspelledWord(sugg2b) {
			if ruleMatch == nil {
				if r.getSpellerFrequency(sugg2a)+r.getSpellerFrequency(sugg2b) > r.getSpellerFrequency(prevWord) {
					ruleMatch = r.createWrongSplitMatch(sentence, ruleMatches, startPos, word, sugg2a, sugg2b, prevStartPos)
					beforeStr = prevWord + " "
				}
			} else {
				// Java: (sugg2a + " " + sugg2b).trim()
				addSug(ruleMatch, tools.JavaStringTrim(sugg2a+" "+sugg2b))
			}
		}
	}

	// "g oing" -> "going"
	sugg := prevWord + word
	if word == strings.ToLower(word) && !r.isMisspelledWord(sugg) {
		if ruleMatch == nil {
			if r.getSpellerFrequency(sugg) >= r.getSpellerFrequency(prevWord) {
				// Java: prevStartPos .. startPos + word.length()
				ruleMatch = spelling.NewSpellingRuleMatch(r, sentence, prevStartPos, startPos+utf16LenMF(word))
				ruleMatch.SetSuggestedReplacement(sugg)
				beforeStr = prevWord + " "
			}
		} else {
			addSug(ruleMatch, sugg)
		}
	}

	if ruleMatch != nil && r.isMisspelledWord(prevWord) {
		return ruleMatch, beforeStr, true
	}
	return ruleMatch, beforeStr, false
}

// tryWrongSplitNext ports getRuleMatches wrong-split with next token.
func (r *MorfologikSpellerRule) tryWrongSplitNext(
	sentence *languagetool.AnalyzedSentence,
	ruleMatches *[]*rules.RuleMatch,
	idx int,
	tokens []*languagetool.AnalyzedTokenReadings,
	word string,
	startPos int,
) (match *rules.RuleMatch, afterStr string, early bool) {
	if r == nil || tokens == nil || word == "" || idx >= len(tokens)-1 {
		return nil, "", false
	}
	nextTok := tokens[idx+1]
	if nextTok == nil || !nextTok.IsWhitespaceBefore() {
		return nil, "", false
	}
	nextWord := nextTok.GetToken()
	if nextWord == "" || strings.ContainsAny(nextWord, digitRunes) {
		return nil, "", false
	}
	if r.getSpellerFrequency(nextWord) >= maxFrequencyForSplitting {
		return nil, "", false
	}
	nextStartPos := nextTok.GetStartPos()
	var ruleMatch *rules.RuleMatch

	if wu := utf16.Encode([]rune(word)); len(wu) >= 1 {
		sugg1a := string(utf16.Decode(wu[:len(wu)-1]))
		sugg1b := string(utf16.Decode(wu[len(wu)-1:])) + nextWord
		if utf16LenMF(sugg1a) > 1 && utf16LenMF(sugg1b) > 2 &&
			!r.isMisspelledWord(sugg1a) && !r.isMisspelledWord(sugg1b) &&
			r.getSpellerFrequency(sugg1a)+r.getSpellerFrequency(sugg1b) > r.getSpellerFrequency(nextWord) {
			ruleMatch = r.createWrongSplitMatch(sentence, ruleMatches, nextStartPos, nextWord, sugg1a, sugg1b, startPos)
			afterStr = " " + nextWord
		}
	}

	if nu := utf16.Encode([]rune(nextWord)); len(nu) > 1 {
		sugg2a := word + string(utf16.Decode(nu[:1]))
		sugg2b := string(utf16.Decode(nu[1:]))
		if utf16LenMF(sugg2b) > 2 && !r.isMisspelledWord(sugg2a) && !r.isMisspelledWord(sugg2b) {
			if ruleMatch == nil {
				if r.getSpellerFrequency(sugg2a)+r.getSpellerFrequency(sugg2b) > r.getSpellerFrequency(nextWord) {
					ruleMatch = r.createWrongSplitMatch(sentence, ruleMatches, nextStartPos, nextWord, sugg2a, sugg2b, startPos)
					afterStr = " " + nextWord
				}
			} else {
				// Java: (sugg2a + " " + sugg2b).trim()
				addSug(ruleMatch, tools.JavaStringTrim(sugg2a+" "+sugg2b))
			}
		}
	}

	sugg := word + nextWord
	if nextWord == strings.ToLower(nextWord) && !r.isMisspelledWord(sugg) {
		if ruleMatch == nil {
			if r.getSpellerFrequency(sugg) >= r.getSpellerFrequency(nextWord) {
				// Java: startPos .. nextStartPos + nextWord.length()
				ruleMatch = spelling.NewSpellingRuleMatch(r, sentence, startPos, nextStartPos+utf16LenMF(nextWord))
				ruleMatch.SetSuggestedReplacement(sugg)
				afterStr = " " + nextWord
			}
		} else {
			addSug(ruleMatch, sugg)
		}
	}

	if ruleMatch != nil && r.isMisspelledWord(nextWord) {
		return ruleMatch, afterStr, true
	}
	return ruleMatch, afterStr, false
}

func addSug(m *rules.RuleMatch, s string) {
	if m == nil || s == "" {
		return
	}
	cur := m.GetSuggestedReplacements()
	for _, x := range cur {
		if x == s {
			return
		}
	}
	m.SetSuggestedReplacements(append(append([]string(nil), cur...), s))
}

func utf16LenMF(s string) int {
	return len(utf16.Encode([]rune(s)))
}
