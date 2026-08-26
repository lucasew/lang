package ga

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/compounds.txt
var compoundsFS embed.FS

// CompoundFilePath is the official resource path Java loads
// (new CompoundRuleData("/ga/compounds.txt") → resource/ga/compounds.txt).
const CompoundFilePath = "/ga/compounds.txt"

// CompoundRuleID ports CompoundRule.getId() → "GA_COMPOUNDS".
const CompoundRuleID = "GA_COMPOUNDS"

var (
	compoundsOnce sync.Once
	compoundsData *rules.CompoundRuleData
)

// loadCompounds ports CompoundRule.getCompoundRuleData lazy singleton.
// Java: new CompoundRuleData("/ga/compounds.txt") via DataBroker →
// resource/ga/compounds.txt (official LT file, embed is byte-identical).
func loadCompounds() *rules.CompoundRuleData {
	compoundsOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, CompoundFilePath)
		if err != nil {
			panic(err)
		}
		compoundsData = d
	})
	return compoundsData
}

// CompoundRule ports org.languagetool.rules.ga.CompoundRule.
// Java extends AbstractCompoundRule with GA hyphen messages and shortMessage;
// no useSubRuleSpecificIds, no anti-patterns, no isMisspelled override, no setUrl,
// no category/issue-type override (keeps AbstractCompoundRule MISC + Misspelling).
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

// NewCompoundRule constructs GA_COMPOUNDS.
// Java: CompoundRule(messages, lang, userConfig) →
//
//	super(messages, lang, userConfig,
//	  "Litrítear an focal seo le fleiscín de ghnáth.",
//	  "Litrítear an focal seo mar fhocal amháin de ghnáth.",
//	  "Litrítear an nath seo mar fhocal amháin nó le fleiscín.",
//	  "Fadhb leis an bhfleiscíniú");
//	addExamplePair(...);
//
// sentenceStartsWithUpperCase remains AbstractCompoundRule default true.
// userConfig null → linguServices null → isMisspelled default false.
func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:          CompoundRuleID,
		Description: "Focail fhleiscínithe, e.g., Moltar 'ró-úsáid' seachas 'ró úsáid'",
		// Java CompoundRule ctor super messages (4-arg including shortMessage):
		WithHyphenMessage:          "Litrítear an focal seo le fleiscín de ghnáth.",
		WithoutHyphenMessage:       "Litrítear an focal seo mar fhocal amháin de ghnáth.",
		WithOrWithoutHyphenMessage: "Litrítear an nath seo mar fhocal amháin nó le fleiscín.",
		ShortDesc:                  "Fadhb leis an bhfleiscíniú",
		// Java AbstractCompoundRule field default true.
		SentenceStartsWithUpperCase: true,
		// Java getCompoundRuleData singleton from /ga/compounds.txt
		Data: loadCompounds(),
	}
	// Java AbstractCompoundRule: Categories.MISC + ITSIssueType.Misspelling.
	rules.InitCompoundRuleMeta(base, messages)
	// Java: mí úsáid → mí-úsáid
	base.AddExamplePair(
		rules.Wrong("Tá <marker>mí úsáid</marker> fhisiciúil i gceist."),
		rules.Fixed("Tá <marker>mí-úsáid</marker> fhisiciúil i gceist."),
	)
	return &CompoundRule{AbstractCompoundRule: base}
}

// GetCompoundRuleData ports CompoundRule.getCompoundRuleData (lazy singleton).
func (r *CompoundRule) GetCompoundRuleData() *rules.CompoundRuleData {
	return loadCompounds()
}

func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
