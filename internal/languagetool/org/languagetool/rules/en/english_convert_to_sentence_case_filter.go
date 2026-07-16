package en

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// EnglishConvertToSentenceCaseFilter ports EnglishConvertToSentenceCaseFilter
// (exception: token "me").
type EnglishConvertToSentenceCaseFilter struct {
	*rules.ConvertToSentenceCaseFilter
}

func NewEnglishConvertToSentenceCaseFilter() *EnglishConvertToSentenceCaseFilter {
	f := rules.NewConvertToSentenceCaseFilter()
	f.TokenIsException = func(s string) bool { return s == "me" }
	return &EnglishConvertToSentenceCaseFilter{ConvertToSentenceCaseFilter: f}
}
