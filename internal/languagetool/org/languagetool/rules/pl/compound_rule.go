package pl

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
// Java: new CompoundRuleData("/pl/compounds.txt") via DataBroker →
// resource/pl/compounds.txt (byte-identical to data/compounds.txt).
func loadCompoundData() *rules.CompoundRuleData {
	compoundOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, "/pl/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// CompoundRule ports org.languagetool.rules.pl.CompoundRule.
// Java extends AbstractCompoundRule with PL hyphen/joined messages and shortMessage;
// example pair Rabce Zdroju → Rabce-Zdroju; no useSubRuleSpecificIds, no anti-patterns,
// no isMisspelled override.
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

// NewCompoundRule constructs PL_COMPOUNDS.
// Java: CompoundRule(messages, lang, userConfig) →
// super(messages, lang, userConfig, withHyphen, withoutHyphen, withOrWithout, shortMessage);
// sentenceStartsWithUpperCase remains AbstractCompoundRule default true.
// userConfig null in Java tests → linguServices null → isMisspelled default false.
func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:          "PL_COMPOUNDS",
		Description: "Sprawdza wyrazy z łącznikiem, np. „łapu capu” zamiast „łapu-capu”",
		// Java CompoundRule ctor super messages (4-arg including shortMessage):
		WithHyphenMessage:          "Ten wyraz pisze się z łącznikiem.",
		WithoutHyphenMessage:       "Ten wyraz pisze się razem (bez spacji ani łącznika).",
		WithOrWithoutHyphenMessage: "Ten wyraz pisze się z łącznikiem lub bez niego.",
		ShortDesc:                  "Brak łącznika lub zbędny łącznik",
		// Java AbstractCompoundRule field default true.
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	// Java AbstractCompoundRule: Categories.MISC + ITSIssueType.Misspelling.
	rules.InitCompoundRuleMeta(base, messages)
	// Java: addExamplePair(wrong("Witamy w <marker>Rabce Zdroju</marker>."),
	//                      fixed("Witamy w <marker>Rabce-Zdroju</marker>."));
	base.AddExamplePair(
		rules.Wrong("Witamy w <marker>Rabce Zdroju</marker>."),
		rules.Fixed("Witamy w <marker>Rabce-Zdroju</marker>."),
	)
	return &CompoundRule{AbstractCompoundRule: base}
}

func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
