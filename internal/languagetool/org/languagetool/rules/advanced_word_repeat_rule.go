package rules

import (
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AdvancedWordRepeatRule ports org.languagetool.rules.AdvancedWordRepeatRule.
// Detects same word/lemma anywhere in the sentence (not only adjacent).
type AdvancedWordRepeatRule struct {
	Messages         map[string]string
	ExcludedWords    map[string]bool
	ExcludedNonWords *regexp.Regexp
	ExcludedPos      *regexp.Regexp
	ID               string
	Message          string
	ShortMessage     string
	// AlsoExcludeSurface skips tokens whose lowercased surface is in ExcludedWords
	// or ExtraSurfaceExcluded (stand-in for prep POS without a tagger).
	AlsoExcludeSurface   bool
	ExtraSurfaceExcluded map[string]bool
}

func (r *AdvancedWordRepeatRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "ADVANCED_WORD_REPEAT"
}

func (r *AdvancedWordRepeatRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	inflectedWords := map[string]bool{}
	curToken := 0
	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		isWord := true
		hasLemma := true

		if len([]rune(token)) < 2 {
			isWord = false
		}

		for _, analyzedToken := range tokens[i].GetReadings() {
			posTag := analyzedToken.GetPOSTag()
			if posTag != nil {
				if tools.IsEmptyStr(*posTag) {
					isWord = false
					break
				}
				lemma := analyzedToken.GetLemma()
				if lemma == nil {
					hasLemma = false
					break
				}
				if r.ExcludedWords[*lemma] {
					isWord = false
					break
				}
				if r.ExcludedPos != nil && r.ExcludedPos.MatchString(*posTag) {
					isWord = false
					break
				}
			} else {
				hasLemma = false
			}
		}

		// Java Matcher.matches() is full-string; Go MatchString is unanchored find.
		if isWord && r.ExcludedNonWords != nil {
			if loc := r.ExcludedNonWords.FindStringIndex(token); loc != nil && loc[0] == 0 && loc[1] == len(token) {
				isWord = false
			}
		}
		if isWord && (r.AlsoExcludeSurface || len(r.ExtraSurfaceExcluded) > 0) {
			tl := strings.ToLower(token)
			if r.ExcludedWords[tl] || r.ExtraSurfaceExcluded[tl] {
				isWord = false
			}
		}

		repetition := false
		prevLemma := ""
		if isWord {
			notSentEnd := false
			for _, analyzedToken := range tokens[i].GetReadings() {
				pos := analyzedToken.GetPOSTag()
				if pos != nil && *pos == languagetool.SentenceEndTagName {
					notSentEnd = true
				}
				if hasLemma {
					lemmaPtr := analyzedToken.GetLemma()
					if lemmaPtr == nil {
						continue
					}
					curLemma := *lemmaPtr
					if prevLemma != curLemma && !notSentEnd {
						if inflectedWords[curLemma] && curToken != i {
							repetition = true
						} else {
							inflectedWords[curLemma] = true
							curToken = i
						}
					}
					prevLemma = curLemma
				} else {
					// Without lemmas, compare case-insensitively (Java lemmas are lowercased).
					key := strings.ToLower(token)
					if !notSentEnd {
						if inflectedWords[key] {
							repetition = true
						} else {
							inflectedWords[key] = true
						}
					}
				}
			}
		}

		if repetition {
			pos := tokens[i].GetStartPos()
			end := pos + utf16LenAdv(token)
			msg := r.Message
			if msg == "" {
				msg = "Word repeated in sentence"
			}
			rm := NewRuleMatch(r, sentence, pos, end, msg)
			rm.ShortMessage = r.ShortMessage
			ruleMatches = append(ruleMatches, rm)
		}
	}
	return ruleMatches
}

func utf16LenAdv(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
