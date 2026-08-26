package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
	chunkxx "github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking/xx"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigxx "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules/xx"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/xx"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// DemoShortCode is Demo.SHORT_CODE ("xx").
const DemoShortCode = "xx"

// Demo ports org.languagetool.language.Demo — always-available test language.
type Demo struct {
	Name string
	// DisambiguationRules optional XML-loaded rules for DemoDisambiguator2.
	DisambiguationRules any // kept untyped to avoid cycles; set via SetDisambiguator
	disambiguator       disambiguation.Disambiguator
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
	if d != nil && d.disambiguator != nil {
		return d.disambiguator
	}
	return disambigxx.NewDemoDisambiguator2()
}

// SetDisambiguator overrides the default DemoDisambiguator2.
func (d *Demo) SetDisambiguator(dis disambiguation.Disambiguator) {
	if d != nil {
		d.disambiguator = dis
	}
}

func (d *Demo) CreateDefaultChunker() chunking.Chunker {
	return chunkxx.NewDemoChunker()
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

// Ensure Demo satisfies languagetool.Language.
var _ languagetool.Language = (*Demo)(nil)
