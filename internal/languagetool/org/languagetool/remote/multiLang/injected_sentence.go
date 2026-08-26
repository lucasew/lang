package multiLang

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"

// InjectedSentence ports org.languagetool.remote.multiLang.InjectedSentence.
type InjectedSentence struct {
	Language string
	Text     string
}

func NewInjectedSentence(language, text string) InjectedSentence {
	// Java constructor stores raw text; only getText() applies String.trim().
	return InjectedSentence{Language: language, Text: text}
}

func (s InjectedSentence) GetLanguage() string { return s.Language }

// GetText ports getText(): return text.trim() (String.trim).
func (s InjectedSentence) GetText() string { return tools.JavaStringTrim(s.Text) }

func (s InjectedSentence) Equal(o InjectedSentence) bool {
	return s.GetLanguage() == o.GetLanguage() && s.GetText() == o.GetText()
}

func (s InjectedSentence) String() string {
	// Java toString uses raw text field.
	return "Sentence: language='" + s.Language + "', text='" + s.Text + "'"
}
