package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/noop"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/xx"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// DemoShortCode is Demo.SHORT_CODE ("xx").
const DemoShortCode = "xx"

// Demo ports org.languagetool.language.Demo — always-available test language.
type Demo struct {
	Name string
}

func NewDemo() *Demo {
	return &Demo{Name: "Testlanguage"}
}

func (d *Demo) GetName() string {
	if d != nil && d.Name != "" {
		return d.Name
	}
	return "Testlanguage"
}

func (d *Demo) GetShortCode() string { return DemoShortCode }

func (d *Demo) GetShortCodeWithCountryAndVariant() string { return DemoShortCode }

func (d *Demo) GetCountries() []string { return []string{"XX"} }

func (d *Demo) GetLocale() string { return "en" }

func (d *Demo) CreateDefaultTagger() languagetool.Tagger {
	return xx.NewDemoTagger()
}

func (d *Demo) CreateDefaultDisambiguator() disambiguation.Disambiguator {
	// Java uses DemoDisambiguator2; identity noop is fine until that port lands.
	return noop.NewNoopDisambiguator()
}

func (d *Demo) CreateDefaultChunker() chunking.Chunker {
	return chunking.FuncChunker(func(_ []*languagetool.AnalyzedTokenReadings) {})
}

func (d *Demo) CreateDefaultWordTokenizer() tokenizers.Tokenizer {
	return tokenizers.NewWordTokenizer()
}

func (d *Demo) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(DemoShortCode)
}

// RegisterDemo registers Demo with the global Languages registry.
func RegisterDemo() {
	languagetool.GlobalLanguages.Register(languagetool.LanguageMeta{
		Name: "Testlanguage",
		Code: DemoShortCode,
	})
}
