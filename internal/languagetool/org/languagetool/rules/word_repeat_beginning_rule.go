package rules

import (
	"regexp"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// WordRepeatBeginningRule ports org.languagetool.rules.WordRepeatBeginningRule.
// Java ctor: setCategory(REPETITIONS_STYLE), setLocQualityIssueType(Style).
type WordRepeatBeginningRule struct {
	Messages map[string]string
	// Category ports Rule.category (Java REPETITIONS_STYLE).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	// Hooks for language subclasses (EnglishWordRepeatBeginningRule).
	IsAdverbFn          func(token *languagetool.AnalyzedTokenReadings) bool
	// IsAdverbAtFn optional context-aware adverb check (e.g. "Sin embargo").
	IsAdverbAtFn        func(tokens []*languagetool.AnalyzedTokenReadings, i int) bool
	IsExceptionFn       func(token string) bool
	IsSentenceException func(sentence *languagetool.AnalyzedSentence) bool
	GetSuggestionsFn    func(token *languagetool.AnalyzedTokenReadings) []string
	IDOverride          string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

func NewWordRepeatBeginningRule(messages map[string]string) *WordRepeatBeginningRule {
	return &WordRepeatBeginningRule{
		Messages:  messages,
		Category:  CatRepetitionsStyle.GetCategory(messages),
		IssueType: ITSStyle,
	}
}

func (r *WordRepeatBeginningRule) GetID() string {
	if r.IDOverride != "" {
		return r.IDOverride
	}
	return "WORD_REPEAT_BEGINNING_RULE"
}

// GetDescription ports WordRepeatBeginningRule.getDescription (messages desc_repetition_beginning).
func (r *WordRepeatBeginningRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["desc_repetition_beginning"]; s != "" {
			return s
		}
	}
	return "Successive sentences beginning with the same word"
}

// GetCategory ports Rule.getCategory.
func (r *WordRepeatBeginningRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// AddExamplePair ports Rule.addExamplePair.
func (r *WordRepeatBeginningRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *WordRepeatBeginningRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *WordRepeatBeginningRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *WordRepeatBeginningRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

// MinToCheckParagraph ports WordRepeatBeginningRule.minToCheckParagraph (Java returns 2).
func (r *WordRepeatBeginningRule) MinToCheckParagraph() int { return 2 }

// EstimateContextForSureMatch ports TextLevelRule (Java always -1).
func (r *WordRepeatBeginningRule) EstimateContextForSureMatch() int { return -1 }

func (r *WordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if r.IsAdverbFn != nil {
		return r.IsAdverbFn(token)
	}
	return false
}

func (r *WordRepeatBeginningRule) isAdverbAt(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if r.IsAdverbAtFn != nil {
		return r.IsAdverbAtFn(tokens, i)
	}
	if i >= 0 && i < len(tokens) {
		return r.isAdverb(tokens[i])
	}
	return false
}

func (r *WordRepeatBeginningRule) isException(token string) bool {
	if r.IsExceptionFn != nil && r.IsExceptionFn(token) {
		return true
	}
	switch token {
	case ":", "–", "-", "✔️", "➡️", "—", "⭐️", "⚠️":
		return true
	}
	return false
}

func (r *WordRepeatBeginningRule) isSentenceException(sentence *languagetool.AnalyzedSentence) bool {
	if r.IsSentenceException != nil {
		return r.IsSentenceException(sentence)
	}
	return false
}

func (r *WordRepeatBeginningRule) getSuggestions(token *languagetool.AnalyzedTokenReadings) []string {
	if r.GetSuggestionsFn != nil {
		return r.GetSuggestionsFn(token)
	}
	return nil
}

var endsSentenceRE = regexp.MustCompile(`.+[.?!]$`)

// MatchList ports match(List<AnalyzedSentence>).
func (r *WordRepeatBeginningRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	lastToken := ""
	beforeLastToken := ""
	var ruleMatches []*RuleMatch
	pos := 0
	var prevSentence *languagetool.AnalyzedSentence
	for _, sentence := range sentences {
		if r.isSentenceException(sentence) {
			prevSentence = nil
			// still advance pos? Java continues without updating lastToken... actually continue after prevSentence=null, still falls through to update? 
			// Java: continue skips the rest of the loop body including beforeLastToken/lastToken/pos updates!
			// So sentence exceptions skip length accumulation too - unusual. Follow Java.
			continue
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		token := ""
		if len(tokens) > 1 {
			analyzedToken := tokens[1]
			token = analyzedToken.GetToken()
			if len(tokens) > 3 {
				// Java: token.length()==1 && !Character.isLetter(charAt(0)) → not a word
				// length/charAt are UTF-16 units (BMP letter = 1 unit; emoji = 2 → stays a word).
				isWord := true
				if utf16LenStr(token) == 1 {
					u := utf16.Encode([]rune(token))
					if len(u) == 1 && !unicode.IsLetter(rune(u[0])) {
						isWord = false
					}
				}
				if isWord && lastToken == token && !r.isException(token) &&
					!r.isException(tokens[2].GetToken()) && !r.isException(tokens[3].GetToken()) &&
					prevSentence != nil &&
					endsSentenceRE.MatchString(stringsTrim(prevSentence.GetText())) {
					var shortMsg string
					if r.isAdverbAt(tokens, 1) {
						shortMsg = r.msg("desc_repetition_beginning_adv", "Adverb repetition at sentence start.")
					} else if beforeLastToken == token {
						shortMsg = r.msg("desc_repetition_beginning_word", "Word repetition at sentence start.")
					}
					if shortMsg != "" {
						thesaurus := r.msg("desc_repetition_beginning_thesaurus", "Consider using a thesaurus.")
						msg := shortMsg + " " + thesaurus
						startPos := analyzedToken.GetStartPos()
						endPos := startPos + utf16LenStr(token)
						rm := NewRuleMatch(r, sentence, pos+startPos, pos+endPos, msg)
						rm.ShortMessage = shortMsg
						suggs := r.getSuggestions(analyzedToken)
						if len(suggs) > 0 {
							rm.SetSuggestedReplacements(suggs)
						}
						ruleMatches = append(ruleMatches, rm)
					}
				}
			}
		}
		beforeLastToken = lastToken
		lastToken = token
		pos += sentence.GetCorrectedTextLength()
		prevSentence = sentence
	}
	return ruleMatches
}

func (r *WordRepeatBeginningRule) msg(key, fallback string) string {
	if r.Messages != nil {
		if s := r.Messages[key]; s != "" {
			return s
		}
	}
	return fallback
}

func stringsTrim(s string) string {
	// strings.TrimSpace
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 {
		c := s[len(s)-1]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	return s
}

func utf16LenStr(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
