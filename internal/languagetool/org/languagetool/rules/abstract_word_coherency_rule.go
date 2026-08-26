package rules

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AbstractWordCoherencyRule ports org.languagetool.rules.AbstractWordCoherencyRule.
// Java ctor: setCategory(MISC). getShortMessage() default null. minToCheckParagraph = -1.
type AbstractWordCoherencyRule struct {
	Messages    map[string]string
	ID          string
	Description string
	WordMap     map[string]map[string]struct{}
	// Category ports Rule.category (Java MISC; languages may override e.g. STYLE).
	Category *Category
	// IssueType ports getLocQualityIssueType (optional; PT sets Inconsistency).
	IssueType ITSIssueType
	// ToBase maps surface form → uninflected file form (lemma stand-in for replacement casing).
	// Not an invent expand of inflections — production loads file pairs only; lemmas come from tagger.
	ToBase map[string]string
	// MessageFn(word1, word2) — word1 is the later variant, word2 the established one.
	MessageFn         func(word1, word2 string) string
	ShortMsg          string // Java getShortMessage(); empty = null
	CreateReplacement func(marked, token, otherSpelling string, tmpToken *languagetool.AnalyzedTokenReadings) string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

// InitWordCoherencyMeta applies Java AbstractWordCoherencyRule constructor metadata.
func InitWordCoherencyMeta(r *AbstractWordCoherencyRule, messages map[string]string) {
	if r == nil {
		return
	}
	r.Messages = messages
	if r.Category == nil {
		r.Category = CatMisc.GetCategory(messages)
	}
}

func (r *AbstractWordCoherencyRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "WORD_COHERENCY"
}

func (r *AbstractWordCoherencyRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	return ""
}

// GetCategory ports Rule.getCategory.
func (r *AbstractWordCoherencyRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *AbstractWordCoherencyRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil {
		return ""
	}
	return r.IssueType
}

// AddExamplePair ports Rule.addExamplePair.
func (r *AbstractWordCoherencyRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AbstractWordCoherencyRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AbstractWordCoherencyRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// MinToCheckParagraph ports AbstractWordCoherencyRule.minToCheckParagraph (Java returns -1).
func (r *AbstractWordCoherencyRule) MinToCheckParagraph() int { return -1 }

// Match ports AbstractWordCoherencyRule.match over sentences.
func (r *AbstractWordCoherencyRule) Match(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	shouldNotAppearWord := make(map[string]string) // later form → established base form
	pos := 0
	for _, sentence := range sentences {
		for _, tmpToken := range sentence.GetTokensWithoutWhitespace() {
			candidates := coherencyCandidates(tmpToken)
			for _, cand := range candidates {
				key := strings.ToLower(cand)
				fromPos := pos + tmpToken.GetStartPos()
				toPos := pos + tmpToken.GetEndPos()
				if other, ok := shouldNotAppearWord[key]; ok {
					msg := r.message(cand, other)
					ruleMatch := NewRuleMatch(r, sentence, fromPos, toPos, msg)
					ruleMatch.ShortMessage = r.ShortMsg
					marked := tmpToken.GetToken()
					// Replace using current base (lemma stand-in) → established base (Java).
					curBase := key
					if r.ToBase != nil {
						if b, ok := r.ToBase[key]; ok {
							curBase = b
						}
					}
					replacement := r.createReplacement(marked, curBase, other, tmpToken)
					if tools.StartsWithUppercase(tmpToken.GetToken()) {
						replacement = tools.UppercaseFirstChar(replacement)
					}
					if !strings.EqualFold(marked, replacement) {
						ruleMatch.SetSuggestedReplacement(replacement)
						ruleMatches = append(ruleMatches, ruleMatch)
					}
					break
				} else if alts, ok := r.WordMap[key]; ok {
					established := key
					if r.ToBase != nil {
						if b, ok := r.ToBase[key]; ok {
							established = b
						}
					}
					for shouldNotAppear := range alts {
						shouldNotAppearWord[shouldNotAppear] = established
					}
				}
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}

func (r *AbstractWordCoherencyRule) message(word1, word2 string) string {
	if r.MessageFn != nil {
		return r.MessageFn(word1, word2)
	}
	return "Do not mix variants of the same word ('" + word1 + "' and '" + word2 + "') within a single text."
}

func (r *AbstractWordCoherencyRule) createReplacement(marked, token, otherSpelling string, tmpToken *languagetool.AnalyzedTokenReadings) string {
	if r.CreateReplacement != nil {
		return r.CreateReplacement(marked, token, otherSpelling, tmpToken)
	}
	return DefaultWordCoherencyReplacement(marked, token, otherSpelling)
}

// DefaultWordCoherencyReplacement ports AbstractWordCoherencyRule.createReplacement
// surface substitution (used when language override falls through).
func DefaultWordCoherencyReplacement(marked, token, otherSpelling string) string {
	re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(token))
	if err != nil {
		return otherSpelling
	}
	loc := re.FindStringIndex(marked)
	if loc == nil {
		// token base not a substring of marked (e.g. reelected vs reelect) — use other directly
		return otherSpelling
	}
	return marked[:loc[0]] + otherSpelling + marked[loc[1]:]
}

func coherencyCandidates(tmpToken *languagetool.AnalyzedTokenReadings) []string {
	var out []string
	seen := map[string]bool{}
	readings := tmpToken.GetReadings()
	if len(readings) == 0 {
		return []string{tmpToken.GetToken()}
	}
	for _, rd := range readings {
		tok := tmpToken.GetToken()
		if rd.GetLemma() != nil && *rd.GetLemma() != "" {
			tok = *rd.GetLemma()
		}
		if !seen[tok] {
			seen[tok] = true
			out = append(out, tok)
		}
	}
	if len(out) == 0 {
		out = []string{tmpToken.GetToken()}
	}
	return out
}
