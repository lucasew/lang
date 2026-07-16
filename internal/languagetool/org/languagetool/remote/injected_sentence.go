package remote

import (
	"fmt"
	"strings"
)

// InjectedSentence ports org.languagetool.remote.multiLang.InjectedSentence.
type InjectedSentence struct {
	Language string
	Text     string
}

func NewInjectedSentence(language, text string) InjectedSentence {
	return InjectedSentence{Language: language, Text: text}
}

func (s InjectedSentence) GetLanguage() string { return s.Language }
func (s InjectedSentence) GetText() string     { return strings.TrimSpace(s.Text) }

func (s InjectedSentence) String() string {
	return fmt.Sprintf("Sentence: language='%s', text='%s'", s.Language, s.GetText())
}

func (s InjectedSentence) Equal(o InjectedSentence) bool {
	return s.Language == o.Language && s.GetText() == o.GetText()
}
