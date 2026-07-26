package ru

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
		// Java: new CompoundRuleData("/ru/compounds.txt")
		d, err := rules.NewCompoundRuleData(f, "/ru/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// RussianCompoundRule ports org.languagetool.rules.ru.RussianCompoundRule.
// Java extends AbstractCompoundRule with RU hyphen/joined messages; no shortMessage,
// no useSubRuleSpecificIds, no anti-patterns, no isMisspelled override.
type RussianCompoundRule struct {
	*rules.AbstractCompoundRule
}

// NewRussianCompoundRule constructs RU_COMPOUNDS.
// Java: RussianCompoundRule(messages, lang, userConfig) →
// super(messages, lang, userConfig, withHyphen, withoutHyphen, withOrWithout);
// addExamplePair; sentenceStartsWithUpperCase = true.
func NewRussianCompoundRule(messages map[string]string) *RussianCompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:          "RU_COMPOUNDS",
		Description: "Правописание через дефис",
		// Java RussianCompoundRule ctor messages (3-arg super → shortDesc null):
		WithHyphenMessage:          "Эти слова должны быть написаны через дефис.",
		WithoutHyphenMessage:       "Эти слова должны быть написаны слитно.",
		WithOrWithoutHyphenMessage: "Эти слова могут быть написаны через дефис или слитно.",
		// Java: super.sentenceStartsWithUpperCase = true (also AbstractCompoundRule default).
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	// Java AbstractCompoundRule: Categories.MISC + ITSIssueType.Misspelling.
	rules.InitCompoundRuleMeta(base, messages)
	// Java: конференц зале → конференц-зале
	base.AddExamplePair(
		rules.Wrong("Собрание состоится в <marker>конференц зале</marker>."),
		rules.Fixed("Собрание состоится в <marker>конференц-зале</marker>."),
	)
	return &RussianCompoundRule{AbstractCompoundRule: base}
}

func (r *RussianCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
