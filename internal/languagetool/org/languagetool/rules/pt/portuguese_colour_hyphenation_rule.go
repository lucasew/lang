package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/compound_colours.txt
var colourCompoundFS embed.FS

// PortugueseColourHyphenationFilePath is the official resource path Java loads
// (CompoundRuleData("/pt/compound_colours.txt")).
const PortugueseColourHyphenationFilePath = "/pt/compound_colours.txt"

// PortugueseColourHyphenationURL ports PortugueseColourHyphenationRule.getUrl()
// → Tools.getUrl("https://pt.wikipedia.org/wiki/Lista_das_alterações_previstas_pelo_acordo_ortográfico_de_1990").
const PortugueseColourHyphenationURL = "https://pt.wikipedia.org/wiki/Lista_das_alterações_previstas_pelo_acordo_ortográfico_de_1990"

var (
	colourCompoundOnce sync.Once
	colourCompoundData *rules.CompoundRuleData
)

func loadColourCompounds() *rules.CompoundRuleData {
	colourCompoundOnce.Do(func() {
		f, err := colourCompoundFS.Open("data/compound_colours.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, PortugueseColourHyphenationFilePath)
		if err != nil {
			panic(err)
		}
		colourCompoundData = d
	})
	return colourCompoundData
}

// PortugueseColourHyphenationRule ports org.languagetool.rules.pt.PortugueseColourHyphenationRule.
type PortugueseColourHyphenationRule struct {
	*rules.AbstractCompoundRule
}

// NewPortugueseColourHyphenationRule constructs PT_COLOUR_HYPHENATION.
// Java ctor: four Portuguese messages to AbstractCompoundRule, then
// setCategory(COMPOUNDING), setLocQualityIssueType(Grammar), useSubRuleSpecificIds().
func NewPortugueseColourHyphenationRule(messages map[string]string) *PortugueseColourHyphenationRule {
	base := &rules.AbstractCompoundRule{
		Messages:                   messages,
		ID:                         "PT_COLOUR_HYPHENATION",
		Description:                "Nomes de cores devem ser hifenizados: \"$match\"",
		WithHyphenMessage:          "Nomes de cores são palavras compostas e devem ser hifenizados.",
		WithoutHyphenMessage:       "Esta palavra é composta por justaposição.",
		WithOrWithoutHyphenMessage: "Esta palavra pode ser composta por justaposição ou hifenizada.",
		ShortDesc:                  "Nomes de cores são palavras compostas.",
		// Java: super.setCategory(Categories.COMPOUNDING.getCategory(messages));
		Category: rules.CatCompounding.GetCategory(messages),
		// Java: setLocQualityIssueType(ITSIssueType.Grammar);
		IssueType: rules.ITSGrammar,
		// Java: getUrl → Tools.getUrl(...)
		URL: PortugueseColourHyphenationURL,
		// Java Language.getShortCode for StringTools.toId (pt-BR → "pt").
		LanguageCode: "pt",
		Data:         loadColourCompounds(),
	}
	// Java: useSubRuleSpecificIds();
	base.UseSubRuleSpecificIDs()
	return &PortugueseColourHyphenationRule{AbstractCompoundRule: base}
}

func (r *PortugueseColourHyphenationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
