package noop

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/xx"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// ShortCode is NoopLanguage.SHORT_CODE ("zz").
const ShortCode = "zz"

// NoopLanguage ports org.languagetool.noop.NoopLanguage — language with no rules.
type NoopLanguage struct {
	Name string
}

func NewNoopLanguage() *NoopLanguage {
	return &NoopLanguage{Name: "NoopLanguage"}
}

func (n *NoopLanguage) GetName() string {
	if n != nil && n.Name != "" {
		return n.Name
	}
	return "NoopLanguage"
}

func (n *NoopLanguage) GetShortCode() string { return ShortCode }

func (n *NoopLanguage) GetCountries() []string { return nil }

func (n *NoopLanguage) CreateDefaultDisambiguator() disambiguation.Disambiguator {
	return NewNoopDisambiguator()
}

func (n *NoopLanguage) CreateDefaultTagger() languagetool.Tagger {
	return xx.NewDemoTagger()
}

func (n *NoopLanguage) CreateDefaultChunker() chunking.Chunker {
	return NewNoopChunker()
}

func (n *NoopLanguage) CreateDefaultWordTokenizer() tokenizers.Tokenizer {
	// Java returns empty list for each call; result must be modifiable.
	return tokenizers.FuncTokenizer(func(text string) []string {
		return []string{}
	})
}

func (n *NoopLanguage) CreateDefaultSentenceTokenizer() tokenizers.Tokenizer {
	return tokenizers.FuncTokenizer(func(text string) []string {
		return []string{text}
	})
}
