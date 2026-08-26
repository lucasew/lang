package rules

import (
	"regexp"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// WordRepeatBeginningRule ports org.languagetool.rules.WordRepeatBeginningRule.
//
// TextLevelRule; category REPETITIONS_STYLE; ITSIssueType Style; minToCheckParagraph=2.
// Default isAdverb/isSentenceException false; getSuggestions empty.
// isException list: : – - ✔️ ➡️ — ⭐️ ⚠️
//
// Match loop (Java bug-for-bug):
//   - isSentenceException → prevSentence=null; continue (skips lastToken/beforeLastToken/pos updates)
//   - tokens = getTokensWithoutWhitespace(); if length>1: token = tokens[1]
//   - if length>3 and isWord and lastToken==token and !isException(token/tokens[2]/tokens[3])
//     and prevSentence!=null and prevSentence.getText().trim().matches(".+[.?!]$"):
//       adverb → shortMsg adv; else if beforeLastToken==token → shortMsg word; else ""
//       if shortMsg non-empty: match on current token span; suggestions from getSuggestions
//   - beforeLastToken=lastToken; lastToken=token; pos+=getCorrectedTextLength(); prevSentence=sentence
//
// Adverb match fires on SECOND consecutive; word match needs THREE (beforeLastToken).
type WordRepeatBeginningRule struct {
	Messages map[string]string
	// Category ports Rule.category (Java REPETITIONS_STYLE).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	// Hooks for language subclasses (e.g. EnglishWordRepeatBeginningRule).
	IsAdverbFn          func(token *languagetool.AnalyzedTokenReadings) bool
	// IsAdverbAtFn optional context-aware adverb check (language modules).
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

// isAdverb ports WordRepeatBeginningRule.isAdverb (default false; subclasses override via IsAdverbFn).
func (r *WordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if r.IsAdverbFn != nil {
		return r.IsAdverbFn(token)
	}
	return false
}

// isAdverbAt is a language-module hook (defaults to isAdverb on tokens[i]).
// Java core only calls isAdverb(analyzedToken); this preserves multi-token subclasses.
func (r *WordRepeatBeginningRule) isAdverbAt(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if r.IsAdverbAtFn != nil {
		return r.IsAdverbAtFn(tokens, i)
	}
	if i >= 0 && i < len(tokens) {
		return r.isAdverb(tokens[i])
	}
	return false
}

// isException ports WordRepeatBeginningRule.isException.
// Subclass IsExceptionFn is OR'd with the base list (Java subclasses call super.isException).
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

// isSentenceException ports WordRepeatBeginningRule.isSentenceException (default false).
func (r *WordRepeatBeginningRule) isSentenceException(sentence *languagetool.AnalyzedSentence) bool {
	if r.IsSentenceException != nil {
		return r.IsSentenceException(sentence)
	}
	return false
}

// getSuggestions ports WordRepeatBeginningRule.getSuggestions (default empty).
func (r *WordRepeatBeginningRule) getSuggestions(token *languagetool.AnalyzedTokenReadings) []string {
	if r.GetSuggestionsFn != nil {
		return r.GetSuggestionsFn(token)
	}
	return nil
}

// endsSentenceRE ports Java prevSentence.getText().trim().matches(".+[.?!]$").
var endsSentenceRE = regexp.MustCompile(`.+[.?!]$`)

// MatchList ports match(List<AnalyzedSentence>).
func (r *WordRepeatBeginningRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	lastToken := ""
	beforeLastToken := ""
	var ruleMatches []*RuleMatch
	pos := 0
	var prevSentence *languagetool.AnalyzedSentence
	for _, sentence := range sentences {
		// Java: if (isSentenceException(sentence)) { prevSentence = null; continue; }
		// continue skips token/lastToken/beforeLastToken/pos updates entirely.
		if r.isSentenceException(sentence) {
			prevSentence = nil
			continue
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		token := ""
		if len(tokens) > 1 {
			analyzedToken := tokens[1]
			token = analyzedToken.GetToken()
			if len(tokens) > 3 {
				// avoid "..." etc. to be matched:
				// Java: isWord = !(token.length()==1 && !Character.isLetter(charAt(0)))
				// length/charAt are UTF-16 units (BMP letter = 1 unit; emoji = 2 → stays a word).
				isWord := true
				if javaStringLen(token) == 1 {
					u := utf16.Encode([]rune(token))
					if len(u) == 1 && !unicode.IsLetter(rune(u[0])) {
						isWord = false
					}
				}
				if isWord && lastToken == token && !r.isException(token) &&
					!r.isException(tokens[2].GetToken()) && !r.isException(tokens[3].GetToken()) &&
					prevSentence != nil &&
					endsSentenceRE.MatchString(tools.JavaStringTrim(prevSentence.GetText())) {
					// Java: isAdverb(analyzedToken); Go also supports IsAdverbAtFn for modules.
					var shortMsg string
					if r.isAdverbAt(tokens, 1) {
						shortMsg = r.msg("desc_repetition_beginning_adv")
					} else if beforeLastToken == token {
						shortMsg = r.msg("desc_repetition_beginning_word")
					} else {
						shortMsg = ""
					}
					if shortMsg != "" {
						msg := shortMsg + " " + r.msg("desc_repetition_beginning_thesaurus")
						// Java: startPos = analyzedToken.getStartPos(); endPos = startPos + token.length()
						startPos := analyzedToken.GetStartPos()
						endPos := startPos + javaStringLen(token)
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

func (r *WordRepeatBeginningRule) msg(key string) string {
	if r.Messages != nil {
		if s := r.Messages[key]; s != "" {
			return s
		}
	}
	// Java ResourceBundle would throw; empty keeps shortMsg gate honest when keys missing.
	return ""
}

// javaStringLen ports Java String.length() (UTF-16 code units).
func javaStringLen(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
