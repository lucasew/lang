package nl

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/compounds.txt
var compoundsFS embed.FS

var (
	compoundOnce sync.Once
	compoundData *rules.CompoundRuleData
)

func loadCompoundData() *rules.CompoundRuleData {
	compoundOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, "/nl/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// CompoundRule ports org.languagetool.rules.nl.CompoundRule.
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:                          "NL_COMPOUNDS",
		Description:                 "Woorden die aaneengeschreven horen, bijvoorbeeld 'zee-egel' i.p.v. 'zee egel': $match",
		WithHyphenMessage:           "Dit woord hoort waarschijnlijk aaneengeschreven met een koppelteken.",
		WithoutHyphenMessage:        "Dit woord hoort waarschijnlijk aaneengeschreven.",
		WithOrWithoutHyphenMessage:  "Deze uitdrukking hoort mogelijk aan elkaar, eventueel met een koppelteken.",
		ShortDesc:                   "Koppeltekenprobleem",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	base.UseSubRuleSpecificIDs()
	rules.InitCompoundRuleMeta(base, messages)
	return &CompoundRule{AbstractCompoundRule: base}
}

func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
