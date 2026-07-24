package en

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// EnglishConvertToSentenceCaseFilter ports
// org.languagetool.rules.en.EnglishConvertToSentenceCaseFilter.
//
// Java only overrides tokenIsException; all acceptRuleMatch logic is core
// ConvertToSentenceCaseFilter. Core calls tokenIsException with the token
// after toLowerCase (and optional typographic-apostrophe rewrite).
type EnglishConvertToSentenceCaseFilter struct {
	*rules.ConvertToSentenceCaseFilter
}

// NewEnglishConvertToSentenceCaseFilter ports the EN subclass constructor
// (implicit super) and wires tokenIsException → s.equals("me").
func NewEnglishConvertToSentenceCaseFilter() *EnglishConvertToSentenceCaseFilter {
	f := rules.NewConvertToSentenceCaseFilter()
	// Java: protected boolean tokenIsException(String s) { return s.equals("me"); }
	f.TokenIsException = englishTokenIsException
	return &EnglishConvertToSentenceCaseFilter{ConvertToSentenceCaseFilter: f}
}

// englishTokenIsException ports EnglishConvertToSentenceCaseFilter.tokenIsException.
func englishTokenIsException(s string) bool {
	return s == "me"
}
