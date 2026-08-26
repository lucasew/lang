package multiLang

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MultiLangCorpora ports org.languagetool.remote.multiLang.MultiLangCorpora.
type MultiLangCorpora struct {
	Language          string
	text              strings.Builder
	InjectedSentences []InjectedSentence
	SentencesInText   int
}

func NewMultiLangCorpora(language string) *MultiLangCorpora {
	return &MultiLangCorpora{Language: language}
}

func (c *MultiLangCorpora) GetLanguage() string { return c.Language }

// GetText ports MultiLangCorpora.getText(): return text.trim() (String.trim).
func (c *MultiLangCorpora) GetText() string {
	if c == nil {
		return ""
	}
	return tools.JavaStringTrim(c.text.String())
}

func (c *MultiLangCorpora) GetInjectedSentences() []InjectedSentence {
	if c == nil {
		return nil
	}
	return append([]InjectedSentence(nil), c.InjectedSentences...)
}

func (c *MultiLangCorpora) GetSentencesInText() int {
	if c == nil {
		return 0
	}
	return c.SentencesInText
}

// InjectOtherSentence appends a foreign-language sentence.
func (c *MultiLangCorpora) InjectOtherSentence(injectLanguage, sentence string) {
	if c == nil {
		return
	}
	c.text.WriteByte(' ')
	c.text.WriteString(sentence)
	c.InjectedSentences = append(c.InjectedSentences, NewInjectedSentence(injectLanguage, sentence))
	c.SentencesInText++
}

// AddSentence appends a main-language sentence.
func (c *MultiLangCorpora) AddSentence(sentence string) {
	if c == nil {
		return
	}
	c.text.WriteByte(' ')
	c.text.WriteString(sentence)
	c.SentencesInText++
}
