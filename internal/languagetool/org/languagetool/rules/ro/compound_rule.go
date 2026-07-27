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

// loadCompoundData ports CompoundRule.getCompoundRuleData lazy singleton.
// Java: new CompoundRuleData("/ro/compounds.txt") via DataBroker →
// resource/ro/compounds.txt (official LT file, embed is byte-identical).
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
// Java extends AbstractCompoundRule with RO hyphen/joined messages and shortMessage;
// no useSubRuleSpecificIds, no anti-patterns, no isMisspelled override, no example pair.
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

// NewCompoundRule constructs RO_COMPOUND.
// Java: CompoundRule(messages, lang, userConfig) →
//
//	super(messages, lang, userConfig,
//	  "Cuvântul se scrie cu cratimă.",
//	  "Cuvântul se scrie legat.",
//	  "Cuvântul se scrie legat sau cu cratimă.",
//	  "Problemă de scriere (cratimă, spațiu, etc.)");
//
// sentenceStartsWithUpperCase remains AbstractCompoundRule default true.
// userConfig null in Java tests → linguServices null → isMisspelled default false.
func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:          "RO_COMPOUND",
		Description: "Greșeală de scriere (cuvinte scrise legat sau cu cratimă)",
		// Java CompoundRule ctor super messages (4-arg including shortMessage):
		WithHyphenMessage:          "Cuvântul se scrie cu cratimă.",
		WithoutHyphenMessage:       "Cuvântul se scrie legat.",
		WithOrWithoutHyphenMessage: "Cuvântul se scrie legat sau cu cratimă.",
		ShortDesc:                  "Problemă de scriere (cratimă, spațiu, etc.)",
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
