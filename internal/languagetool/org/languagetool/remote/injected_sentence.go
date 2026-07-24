package remote

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// InjectedSentence ports org.languagetool.remote.multiLang.InjectedSentence.
type InjectedSentence struct {
	Language string
	Text     string
}

func NewInjectedSentence(language, text string) InjectedSentence {
	// Java stores raw text; only getText() trims.
	return InjectedSentence{Language: language, Text: text}
}

func (s InjectedSentence) GetLanguage() string { return s.Language }

// GetText ports getText(): return text.trim() (String.trim, not Unicode TrimSpace).
func (s InjectedSentence) GetText() string { return tools.JavaStringTrim(s.Text) }

func (s InjectedSentence) String() string {
	// Java toString uses the raw text field (not getText() / trim).
	return fmt.Sprintf("Sentence: language='%s', text='%s'", s.Language, s.Text)
}

func (s InjectedSentence) Equal(o InjectedSentence) bool {
	return s.Language == o.Language && s.GetText() == o.GetText()
}
