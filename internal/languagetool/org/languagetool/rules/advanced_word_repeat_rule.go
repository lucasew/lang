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
// Abstract hooks (Java): getExcludedWordsPattern, getExcludedNonWordsPattern,
// getExcludedPos, getMessage, getShortMessage — fields on this struct.
type AdvancedWordRepeatRule struct {
	Messages         map[string]string
	ExcludedWords    map[string]bool // Java getExcludedWordsPattern() Set
	ExcludedNonWords *regexp.Regexp  // Java getExcludedNonWordsPattern()
	ExcludedPos      *regexp.Regexp  // Java getExcludedPos()
	ID               string
	Message          string // Java getMessage()
	ShortMessage     string // Java getShortMessage()
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

// Match ports AdvancedWordRepeatRule.match(AnalyzedSentence).
// Control flow is a line-faithful transcription of the Java method.
func (r *AdvancedWordRepeatRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	repetition := false
	// Java: Set<String> inflectedWords = new TreeSet<>() — contains/add only.
	inflectedWords := map[string]bool{}
	curToken := 0
	// start from real token, 0 = SENT_START
	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		// avoid "..." etc. to be matched:
		isWord := true
		hasLemma := true

		// Java: token.length() < 2 (UTF-16 code units)
		if utf16LenAdv(token) < 2 {
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

				// Java: Matcher.matches() ≡ whole-string match
				if javaRegexpMatches(r.ExcludedPos, *posTag) {
					isWord = false
					break
				}
			} else {
				hasLemma = false
			}
		}

		// Java: Matcher m1 = getExcludedNonWordsPattern().matcher(token); m1.matches()
		if isWord && javaRegexpMatches(r.ExcludedNonWords, token) {
			isWord = false
		}

		prevLemma := ""
		if isWord {
			// Java misnames notSentEnd: true when any reading is SENT_END.
			notSentEnd := false
			for _, analyzedToken := range tokens[i].GetReadings() {
				pos := analyzedToken.GetPOSTag()
				if pos != nil {
					notSentEnd = notSentEnd || *pos == languagetool.SentenceEndTagName
				}
				if hasLemma {
					// When hasLemma, Java first loop established non-null lemmas on non-null POS readings.
					curLemmaPtr := analyzedToken.GetLemma()
					var curLemma string
					if curLemmaPtr != nil {
						curLemma = *curLemmaPtr
					}
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
					if inflectedWords[token] && !notSentEnd {
						repetition = true
					} else {
						inflectedWords[token] = true
					}
				}
			}
		}

		if repetition {
			pos := tokens[i].GetStartPos()
			// Java: pos + token.length() (UTF-16)
			ruleMatch := NewRuleMatch(r, sentence, pos, pos+utf16LenAdv(token), r.Message)
			ruleMatch.ShortMessage = r.ShortMessage
			ruleMatches = append(ruleMatches, ruleMatch)
			repetition = false
		}
	}
	return ruleMatches
}

// javaRegexpMatches ports java.util.regex.Matcher.matches(): whole-string match.
// Go regexp.MatchString is unanchored find (≈ Matcher.find()).
func javaRegexpMatches(re *regexp.Regexp, s string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}

func utf16LenAdv(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
