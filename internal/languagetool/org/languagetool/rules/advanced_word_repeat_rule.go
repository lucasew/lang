package rules

import (
	"regexp"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AdvancedWordRepeatRule ports org.languagetool.rules.AdvancedWordRepeatRule.
// Detects same word/lemma anywhere in the sentence (not only adjacent).
// Java ctor: setCategory(MISC), setDefaultOff(), setLocQualityIssueType(Style).
// ExcludedWords match lemmas only (Java getExcludedWordsPattern); prep/pronoun
// exclusions use ExcludedPos when tagged — no surface invent of POS lists.
type AdvancedWordRepeatRule struct {
	Messages         map[string]string
	ExcludedWords    map[string]bool
	ExcludedNonWords *regexp.Regexp
	ExcludedPos      *regexp.Regexp
	ID               string
	Message          string
	ShortMessage     string
	// Category ports Rule.category (Java MISC).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	// DefaultOff ports setDefaultOff() (Java true for this abstract rule).
	DefaultOff bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

// AddExamplePair ports Rule.addExamplePair.
func (r *AdvancedWordRepeatRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AdvancedWordRepeatRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AdvancedWordRepeatRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// InitAdvancedWordRepeatMeta applies Java AdvancedWordRepeatRule constructor metadata.
func InitAdvancedWordRepeatMeta(r *AdvancedWordRepeatRule, messages map[string]string) {
	if r == nil {
		return
	}
	r.Messages = messages
	if r.Category == nil {
		r.Category = CatMisc.GetCategory(messages)
	}
	if r.IssueType == "" {
		r.IssueType = ITSStyle
	}
	r.DefaultOff = true
}

func (r *AdvancedWordRepeatRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "ADVANCED_WORD_REPEAT"
}

// GetCategory ports Rule.getCategory.
func (r *AdvancedWordRepeatRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *AdvancedWordRepeatRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

// IsDefaultOff ports Rule.isDefaultOff.
func (r *AdvancedWordRepeatRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
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

		// Java: boolean notSentEnd misnamed — true when any reading is SENT_END.
		repetition := false
		prevLemma := ""
		if isWord {
			// Soft Analyze may leave only SENT_END on the last content word (null-POS
			// reading dropped). Still track surface so untagged "test test" matches
			// (incomplete vs full Java tagger, not surface invent of POS/lemma lists).
			contentReadings := 0
			isSentEnd := false
			for _, analyzedToken := range tokens[i].GetReadings() {
				pos := analyzedToken.GetPOSTag()
				if pos != nil && *pos == languagetool.SentenceEndTagName {
					isSentEnd = true
				}
				if pos != nil && (*pos == languagetool.SentenceEndTagName || *pos == languagetool.ParagraphEndTagName) {
					continue
				}
				contentReadings++
				if hasLemma {
					lemmaPtr := analyzedToken.GetLemma()
					if lemmaPtr == nil {
						continue
					}
					curLemma := *lemmaPtr
					// Java: if (!prevLemma.equals(curLemma) && !notSentEnd)
					if prevLemma != curLemma && !isSentEnd {
						if inflectedWords[curLemma] && curToken != i {
							repetition = true
						} else {
							inflectedWords[curLemma] = true
							curToken = i
						}
					}
					prevLemma = curLemma
				} else {
					// Java: inflectedWords.contains(tokens[i].getToken()) — exact surface.
					if !isSentEnd {
						if inflectedWords[token] {
							repetition = true
						} else {
							inflectedWords[token] = true
						}
					}
				}
			}
			if contentReadings == 0 && !hasLemma {
				if inflectedWords[token] {
					repetition = true
				} else {
					inflectedWords[token] = true
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
			// Java resets repetition after each match.
			repetition = false
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
