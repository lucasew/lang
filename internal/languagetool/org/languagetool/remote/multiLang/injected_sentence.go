package multiLang

import "strings"

// InjectedSentence ports org.languagetool.remote.multiLang.InjectedSentence.
type InjectedSentence struct {
	Language string
	Text     string
}

func NewInjectedSentence(language, text string) InjectedSentence {
	return InjectedSentence{Language: language, Text: strings.TrimSpace(text)}
}

func (s InjectedSentence) GetLanguage() string { return s.Language }
func (s InjectedSentence) GetText() string     { return strings.TrimSpace(s.Text) }

func (s InjectedSentence) Equal(o InjectedSentence) bool {
	return s.GetLanguage() == o.GetLanguage() && s.GetText() == o.GetText()
}

func (s InjectedSentence) String() string {
	return "Sentence: language='" + s.Language + "', text='" + s.GetText() + "'"
}
