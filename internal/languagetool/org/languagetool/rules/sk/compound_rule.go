package sk

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
		d, err := rules.NewCompoundRuleData(f, "/sk/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// CompoundRule ports org.languagetool.rules.sk.CompoundRule.
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                    messages,
		ID:                          "SK_COMPOUNDS",
		Description:                 "Slová so spojovníkom napr. použite „česko-slovenský” namiesto „česko slovenský”",
		WithHyphenMessage:           "Toto slovo sa zvyčajne píše so spojovníkom.",
		WithoutHyphenMessage:        "Toto slovo sa obvykle píše bez spojovníka.",
		WithOrWithoutHyphenMessage:  "Tento výraz sa bežne píše s alebo bez spojovníka.",
		ShortDesc:                   "Problém spájania slov",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	return &CompoundRule{AbstractCompoundRule: base}
}

func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
