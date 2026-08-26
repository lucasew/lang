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

// loadCompoundData ports CompoundRule.getCompoundRuleData lazy singleton.
// Java: new CompoundRuleData("/sk/compounds.txt")
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
// Java extends AbstractCompoundRule with SK hyphen messages; shortMessage set;
// no useSubRuleSpecificIds, no anti-patterns, no isMisspelled override, no example pair.
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

// NewCompoundRule constructs SK_COMPOUNDS.
// Java: CompoundRule(messages, lang, userConfig) →
// super(messages, lang, userConfig, withHyphen, withoutHyphen, withOrWithout, shortMessage);
// sentenceStartsWithUpperCase remains AbstractCompoundRule default true.
// userConfig null in Java tests → linguServices null → isMisspelled default false.
func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:          "SK_COMPOUNDS",
		Description: "Slová so spojovníkom napr. použite „česko-slovenský” namiesto „česko slovenský”",
		// Java CompoundRule ctor super messages (4-arg including shortMessage):
		WithHyphenMessage:          "Toto slovo sa zvyčajne píše so spojovníkom.",
		WithoutHyphenMessage:       "Toto slovo sa obvykle píše bez spojovníka.",
		WithOrWithoutHyphenMessage: "Tento výraz sa bežne píše s alebo bez spojovníka.",
		ShortDesc:                  "Problém spájania slov",
		// Java AbstractCompoundRule field default true.
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	// Java AbstractCompoundRule: Categories.MISC + ITSIssueType.Misspelling.
	rules.InitCompoundRuleMeta(base, messages)
	return &CompoundRule{AbstractCompoundRule: base}
}

func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
