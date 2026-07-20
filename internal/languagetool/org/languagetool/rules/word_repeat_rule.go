package rules

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// WordRepeatRule ports org.languagetool.rules.WordRepeatRule.
// Java ctor: setCategory(MISC), setLocQualityIssueType(Duplication).
type WordRepeatRule struct {
	Messages map[string]string
	// Category ports Rule.category (Java MISC).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Duplication).
	IssueType ITSIssueType
	// ExtraIgnore is called from Ignore for language-specific exceptions (e.g. EnglishWordRepeatRule).
	ExtraIgnore func(tokens []*languagetool.AnalyzedTokenReadings, position int) bool
	// CreateMatchFn optional override for createRuleMatch (e.g. Ukrainian І/і suggestion).
	CreateMatchFn func(r *WordRepeatRule, sentence *languagetool.AnalyzedSentence, prevToken, token string, prevPos, pos int, msg string) *RuleMatch
	// IDOverride when non-empty replaces the default WORD_REPEAT_RULE id.
	IDOverride string
	// AntiPatterns ports Rule.getAntiPatterns (IMMUNIZE/IGNORE_SPELLING via Replace).
	// Used by Match via SentenceWithImmunization (Java WordRepeatRule.match).
	AntiPatterns []SentenceReplacer
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

func NewWordRepeatRule(messages map[string]string) *WordRepeatRule {
	return &WordRepeatRule{
		Messages:  messages,
		Category:  CatMisc.GetCategory(messages),
		IssueType: ITSDuplication,
	}
}

func (r *WordRepeatRule) GetID() string {
	if r.IDOverride != "" {
		return r.IDOverride
	}
	return "WORD_REPEAT_RULE"
}

// GetDescription ports WordRepeatRule.getDescription (messages "desc_repetition").
func (r *WordRepeatRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["desc_repetition"]; s != "" {
			return s
		}
	}
	return "Word repetition (e.g. 'will will')"
}

// GetCategory ports Rule.getCategory.
func (r *WordRepeatRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *WordRepeatRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSDuplication
	}
	return r.IssueType
}

// AddExamplePair ports Rule.addExamplePair.
func (r *WordRepeatRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *WordRepeatRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *WordRepeatRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// EstimateContextForSureMatch ports estimateContextForSureMatch (Java returns 1).
func (r *WordRepeatRule) EstimateContextForSureMatch() int { return 1 }

func (r *WordRepeatRule) shortMessage() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["desc_repetition_short"]; s != "" {
			return s
		}
	}
	return "Word repetition"
}

func (r *WordRepeatRule) Ignore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if r.ExtraIgnore != nil && r.ExtraIgnore(tokens, position) {
		return true
	}
	for _, w := range []string{"Phi", "Li", "Xiao", "Duran", "Wagga", "Abdullah", "Nwe", "Pago", "Cao"} {
		if r.wordRepetitionOf(w, tokens, position) {
			return true
		}
	}
	return false
}

func (r *WordRepeatRule) wordRepetitionOf(word string, tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	return position > 0 &&
		tokens[position-1].GetToken() == word &&
		tokens[position].GetToken() == word
}

func (r *WordRepeatRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	// Java: getSentenceWithImmunization(sentence).getTokensWithoutWhitespace()
	work := sentence
	if r != nil {
		work = SentenceWithImmunization(sentence, r.AntiPatterns)
	}
	tokens := work.GetTokensWithoutWhitespace()
	prevToken := ""
	msg := r.Messages["repetition"]
	if msg == "" {
		msg = "Word repetition"
	}
	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		if tokens[i].IsImmunized() {
			prevToken = ""
			continue
		}
		if isWord(token) && strings.EqualFold(prevToken, token) && !r.Ignore(tokens, i) {
			prevPos := tokens[i-1].GetStartPos()
			pos := tokens[i].GetStartPos()
			var rm *RuleMatch
			if r.CreateMatchFn != nil {
				rm = r.CreateMatchFn(r, sentence, prevToken, token, prevPos, pos, msg)
			} else {
				// Java createRuleMatch: shortMessage = messages.getString("desc_repetition_short")
				rm = NewRuleMatch(r, sentence, prevPos, pos+utf16Len(prevToken), msg)
				rm.ShortMessage = r.shortMessage()
				rm.SetSuggestedReplacement(prevToken)
			}
			ruleMatches = append(ruleMatches, rm)
		}
		prevToken = token
	}
	return ruleMatches
}

func isWord(token string) bool {
	if tools.IsEmoji(token) {
		return false
	}
	if tools.IsNumericSpace(token) {
		return false
	}
	runes := []rune(token)
	if len(runes) == 1 {
		if !unicode.IsLetter(runes[0]) {
			return false
		}
	}
	return true
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}
