package ga

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/compounds.txt
var compoundsFS embed.FS

var (
	compoundsOnce sync.Once
	compoundsData *rules.CompoundRuleData
)

func loadCompounds() *rules.CompoundRuleData {
	compoundsOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, "/ga/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundsData = d
	})
	return compoundsData
}

// CompoundRule ports org.languagetool.rules.ga.CompoundRule.
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:                         "GA_COMPOUNDS",
		Description:                "Focail fhleiscínithe, e.g., Moltar 'ró-úsáid' seachas 'ró úsáid'",
		WithHyphenMessage:          "Litrítear an focal seo le fleiscín de ghnáth.",
		WithoutHyphenMessage:       "Litrítear an focal seo mar fhocal amháin de ghnáth.",
		WithOrWithoutHyphenMessage: "Litrítear an nath seo mar fhocal amháin nó le fleiscín.",
		ShortDesc:                  "Fadhb leis an bhfleiscíniú",
		Data:                       loadCompounds(),
	}
	rules.InitCompoundRuleMeta(base, messages)
	// Java: mí úsáid → mí-úsáid
	base.AddExamplePair(
		rules.Wrong("Tá <marker>mí úsáid</marker> fhisiciúil i gceist."),
		rules.Fixed("Tá <marker>mí-úsáid</marker> fhisiciúil i gceist."),
	)
	return &CompoundRule{AbstractCompoundRule: base}
}

func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
