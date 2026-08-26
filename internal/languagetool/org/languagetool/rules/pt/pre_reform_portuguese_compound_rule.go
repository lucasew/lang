package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/pre-reform-compounds.txt
var preReformFS embed.FS

// PreReformPortugueseCompoundFilePath is the official resource path Java loads
// (CompoundRuleData("/pt/pre-reform-compounds.txt")).
const PreReformPortugueseCompoundFilePath = "/pt/pre-reform-compounds.txt"

var (
	preReformOnce sync.Once
	preReformData *rules.CompoundRuleData
)

func loadPreReformCompounds() *rules.CompoundRuleData {
	preReformOnce.Do(func() {
		f, err := preReformFS.Open("data/pre-reform-compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, PreReformPortugueseCompoundFilePath)
		if err != nil {
			panic(err)
		}
		preReformData = d
	})
	return preReformData
}

// PreReformPortugueseCompoundRule ports org.languagetool.rules.pt.PreReformPortugueseCompoundRule.
type PreReformPortugueseCompoundRule struct {
	*rules.AbstractCompoundRule
}

// NewPreReformPortugueseCompoundRule constructs PT_COMPOUNDS_PRE_REFORM.
// Java ctor: four Portuguese messages to AbstractCompoundRule, then
// setCategory(COMPOUNDING), setLocQualityIssueType(Grammar), useSubRuleSpecificIds().
// No getUrl override (empty URL). No mergeCompound override (default base only).
func NewPreReformPortugueseCompoundRule(messages map[string]string) *PreReformPortugueseCompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                   messages,
		ID:                         "PT_COMPOUNDS_PRE_REFORM",
		Description:                "Palavras compostas",
		WithHyphenMessage:          "Esta palavra é hifenizada.",
		WithoutHyphenMessage:       "Esta palavra é composta por justaposição.",
		WithOrWithoutHyphenMessage: "Esta palavra pode ser composta por justaposição ou hifenizada.",
		ShortDesc:                  "Este conjunto forma uma palavra composta.",
		// Java: super.setCategory(Categories.COMPOUNDING.getCategory(messages));
		Category: rules.CatCompounding.GetCategory(messages),
		// Java: setLocQualityIssueType(ITSIssueType.Grammar);
		IssueType: rules.ITSGrammar,
		// Java: no getUrl override → empty URL
		// Java Language.getShortCode for StringTools.toId (pt-BR → "pt").
		LanguageCode: "pt",
		// Java AbstractCompoundRule.sentenceStartsWithUpperCase default true
		// (needed so "Guarda chuva" matches data entry "guarda chuva").
		SentenceStartsWithUpperCase: true,
		Data:                        loadPreReformCompounds(),
	}
	// Java: no mergeCompound override → MergeCompoundFn left nil (default base).
	// Java: this.useSubRuleSpecificIds();
	base.UseSubRuleSpecificIDs()
	return &PreReformPortugueseCompoundRule{AbstractCompoundRule: base}
}

func (r *PreReformPortugueseCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
