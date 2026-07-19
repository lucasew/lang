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
// Java ctor: setCategory(MISC), setDefaultOff(), setLocQualityIssueType(Style).
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
	// AlsoExcludeSurface skips tokens whose lowercased surface is in ExcludedWords
	// or ExtraSurfaceExcluded (stand-in for prep POS without a tagger).
	AlsoExcludeSurface   bool
	ExtraSurfaceExcluded map[string]bool
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
		if isWord && (r.AlsoExcludeSurface || len(r.ExtraSurfaceExcluded) > 0) {
			tl := strings.ToLower(token)
			if r.ExcludedWords[tl] || r.ExtraSurfaceExcluded[tl] {
				isWord = false
			}
		}

		repetition := false
		prevLemma := ""
		if isWord {
			// Soft Analyze uses AddReading for SENT_END, which (like Java) drops a
			// trailing null-POS reading. The last content word then has only SENT_END.
			// Still count its surface so two-word sentences without "." match
			// (server multi-lang "test test" / PL_WORD_REPEAT).
			contentReadings := 0
			for _, analyzedToken := range tokens[i].GetReadings() {
				pos := analyzedToken.GetPOSTag()
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
					if prevLemma != curLemma {
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
					if inflectedWords[key] {
						repetition = true
					} else {
						inflectedWords[key] = true
					}
				}
			}
			if contentReadings == 0 && !hasLemma {
				// Pure SENT_END annotation on last content word: still track surface.
				key := strings.ToLower(token)
				if inflectedWords[key] {
					repetition = true
				} else {
					inflectedWords[key] = true
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
