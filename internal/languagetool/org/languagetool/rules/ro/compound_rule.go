package ro

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
		d, err := rules.NewCompoundRuleData(f, "/ro/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// CompoundRule ports org.languagetool.rules.ro.CompoundRule.
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                    messages,
		ID:                          "RO_COMPOUND",
		Description:                 "Greșeală de scriere (cuvinte scrise legat sau cu cratimă)",
		WithHyphenMessage:           "Cuvântul se scrie cu cratimă.",
		WithoutHyphenMessage:        "Cuvântul se scrie legat.",
		WithOrWithoutHyphenMessage:  "Cuvântul se scrie legat sau cu cratimă.",
		ShortDesc:                   "Problemă de scriere (cratimă, spațiu, etc.)",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	return &CompoundRule{AbstractCompoundRule: base}
}

func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
