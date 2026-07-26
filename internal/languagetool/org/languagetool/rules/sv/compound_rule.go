package sv

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
// Java: new CompoundRuleData("/sv/compounds.txt")
func loadCompoundData() *rules.CompoundRuleData {
	compoundOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, "/sv/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// CompoundRule ports org.languagetool.rules.sv.CompoundRule.
// Java extends AbstractCompoundRule with SV hyphen messages (3-arg super, no shortMessage);
// no useSubRuleSpecificIds, no anti-patterns, no isMisspelled override, no example pair.
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

// NewCompoundRule constructs SV_COMPOUNDS.
// Java: CompoundRule(messages, lang, userConfig) →
// super(messages, lang, userConfig, withHyphen, withoutHyphen, withOrWithout);
// sentenceStartsWithUpperCase remains AbstractCompoundRule default true.
// userConfig null in Java tests → linguServices null → isMisspelled default false.
func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:          "SV_COMPOUNDS",
		Description: "Särskrivningar, t.ex. 'e mail' bör skrivas 'e-mail'",
		// Java CompoundRule ctor super messages (3-arg, shortMessage null):
		WithHyphenMessage:          "Dessa ord skrivs samman med bindestreck.",
		WithoutHyphenMessage:       "Dessa ord skrivs samman.",
		WithOrWithoutHyphenMessage: "Dessa ord skrivs samman med eller utan bindestreck.",
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
