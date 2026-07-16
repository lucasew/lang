package languagetool

// SimpleInputSentence ports org.languagetool.SimpleInputSentence — cache key for analysis.
type SimpleInputSentence struct {
	Text         string
	LanguageCode string
}

func NewSimpleInputSentence(text, languageCode string) SimpleInputSentence {
	if languageCode == "" {
		panic("language required")
	}
	return SimpleInputSentence{Text: text, LanguageCode: languageCode}
}

func (s SimpleInputSentence) GetText() string { return s.Text }

func (s SimpleInputSentence) Equal(o SimpleInputSentence) bool {
	return s.Text == o.Text && s.LanguageCode == o.LanguageCode
}

func (s SimpleInputSentence) String() string { return s.Text }
