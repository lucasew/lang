package rules

import (
	"regexp"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// WordRepeatBeginningRule ports org.languagetool.rules.WordRepeatBeginningRule.
type WordRepeatBeginningRule struct {
	Messages map[string]string
	// Hooks for language subclasses (EnglishWordRepeatBeginningRule).
	IsAdverbFn          func(token *languagetool.AnalyzedTokenReadings) bool
	IsExceptionFn       func(token string) bool
	IsSentenceException func(sentence *languagetool.AnalyzedSentence) bool
	GetSuggestionsFn    func(token *languagetool.AnalyzedTokenReadings) []string
	IDOverride          string
}

func NewWordRepeatBeginningRule(messages map[string]string) *WordRepeatBeginningRule {
	return &WordRepeatBeginningRule{Messages: messages}
}

func (r *WordRepeatBeginningRule) GetID() string {
	if r.IDOverride != "" {
		return r.IDOverride
	}
	return "WORD_REPEAT_BEGINNING_RULE"
}

func (r *WordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if r.IsAdverbFn != nil {
		return r.IsAdverbFn(token)
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
				isWord := true
				runes := []rune(token)
				if len(runes) == 1 && !unicode.IsLetter(runes[0]) {
					isWord = false
				}
				if isWord && lastToken == token && !r.isException(token) &&
					!r.isException(tokens[2].GetToken()) && !r.isException(tokens[3].GetToken()) &&
					prevSentence != nil &&
					endsSentenceRE.MatchString(stringsTrim(prevSentence.GetText())) {
					var shortMsg string
					if r.isAdverb(analyzedToken) {
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
