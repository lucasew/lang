package tokenizers

// EnglishSRXSentenceTokenizer ports tokenizers.EnglishSRXSentenceTokenizer.
type EnglishSRXSentenceTokenizer struct {
	*SRXSentenceTokenizer
}

func NewEnglishSRXSentenceTokenizer() *EnglishSRXSentenceTokenizer {
	return &EnglishSRXSentenceTokenizer{
		SRXSentenceTokenizer: NewSRXSentenceTokenizer("en"),
	}
}
