package br

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/compounds.txt
var compoundsFS embed.FS

// BretonCompoundFilePath is the official resource path Java loads
// (new CompoundRuleData("/br/compounds.txt")).
const BretonCompoundFilePath = "/br/compounds.txt"

var (
	compoundOnce sync.Once
	compoundData *rules.CompoundRuleData
)

// loadCompoundData ports BretonCompoundRule.getCompoundRuleData lazy singleton.
// Java: new CompoundRuleData("/br/compounds.txt") via DataBroker →
// resource/br/compounds.txt (official LT file, embed is byte-identical).
func loadCompoundData() *rules.CompoundRuleData {
	compoundOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, BretonCompoundFilePath)
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// BretonCompoundRule ports org.languagetool.rules.br.BretonCompoundRule.
// Java extends AbstractCompoundRule with BR hyphen messages; then overrides
// category to COMPOUNDING and ITSIssueType to Grammar (not MISC/Misspelling defaults).
// No useSubRuleSpecificIds, no anti-patterns, no isMisspelled override, no getUrl.
type BretonCompoundRule struct {
	*rules.AbstractCompoundRule
}

// NewBretonCompoundRule constructs BR_COMPOUNDS.
// Java: BretonCompoundRule(messages, lang, userConfig) →
//
//	super(messages, lang, userConfig,
//	  "Skrivet e vez ar ger-mañ boaz gant ur varrennig-stagañ.",
//	  "Ar ger-mañ a zo skrivet boaz evel unan hepken.",
//	  "An droienn-mañ a zo skrivet evel ur ger hepken pe gant ur varrennig-stagañ.",
//	  "Kudenn barrennig-stagañ");
//	super.setCategory(Categories.COMPOUNDING.getCategory(messages));
//	addExamplePair(...);
//	setLocQualityIssueType(ITSIssueType.Grammar);
//
// sentenceStartsWithUpperCase remains AbstractCompoundRule default true.
// userConfig null → linguServices null → isMisspelled default false.
func NewBretonCompoundRule(messages map[string]string) *BretonCompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:    messages,
		ID:          "BR_COMPOUNDS",
		Description: "Mots composés",
		// Java 4-arg super messages (including shortMessage):
		WithHyphenMessage:          "Skrivet e vez ar ger-mañ boaz gant ur varrennig-stagañ.",
		WithoutHyphenMessage:       "Ar ger-mañ a zo skrivet boaz evel unan hepken.",
		WithOrWithoutHyphenMessage: "An droienn-mañ a zo skrivet evel ur ger hepken pe gant ur varrennig-stagañ.",
		ShortDesc:                  "Kudenn barrennig-stagañ",
		// Java: super.setCategory(Categories.COMPOUNDING.getCategory(messages));
		// (overrides AbstractCompoundRule default MISC — do not use InitCompoundRuleMeta alone)
		Category: rules.CatCompounding.GetCategory(messages),
		// Java: setLocQualityIssueType(ITSIssueType.Grammar);
		// (overrides AbstractCompoundRule default Misspelling)
		IssueType: rules.ITSGrammar,
		// Java AbstractCompoundRule field default true.
		SentenceStartsWithUpperCase: true,
		// Java getCompoundRuleData singleton from /br/compounds.txt
		Data: loadCompoundData(),
	}
	// Java: alc'hweder gwez → alc'hweder-gwez
	base.AddExamplePair(
		rules.Wrong("Gwelet em eus un <marker>alc'hweder gwez</marker> e-kerzh an dibenn-sizhun-mañ."),
		rules.Fixed("Gwelet em eus un <marker>alc'hweder-gwez</marker> e-kerzh an dibenn-sizhun-mañ."),
	)
	return &BretonCompoundRule{AbstractCompoundRule: base}
}

// GetCompoundRuleData ports BretonCompoundRule.getCompoundRuleData (lazy singleton).
func (r *BretonCompoundRule) GetCompoundRuleData() *rules.CompoundRuleData {
	return loadCompoundData()
}

func (r *BretonCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
