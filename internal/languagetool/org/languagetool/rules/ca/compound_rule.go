package ca

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
		d, err := rules.NewCompoundRuleData(f, "/ca/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// CompoundRule ports org.languagetool.rules.ca.CompoundRule.
// isMisspelled uses TagIsTagged (Java CatalanTagger.INSTANCE_VAL.tag(...).isTagged()).
// Without TagIsTagged, isMisspelled is false (AbstractCompoundRule default).
type CompoundRule struct {
	*rules.AbstractCompoundRule
	// TagIsTagged ports CatalanTagger.tag([word])[0].isTagged(); nil → misspelled=false.
	TagIsTagged func(word string) bool
}

func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:                          "CA_COMPOUNDS",
		Description:                 "Paraules compostes amb guionet: $match",
		WithHyphenMessage:           "S'escriu amb un guionet.",
		WithoutHyphenMessage:        "S'escriu junt sense espai ni guionet.",
		WithOrWithoutHyphenMessage:  "S'escriu junt o amb guionet.",
		ShortDesc:                   "Error de mot compost",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	base.UseSubRuleSpecificIDs()
	rules.InitCompoundRuleMeta(base, messages)
	r := &CompoundRule{AbstractCompoundRule: base}
	base.IsMisspelled = func(word string) bool {
		if r.TagIsTagged == nil {
			return false
		}
		return !r.TagIsTagged(word)
	}
	return r
}

// WireCompoundRuleTagger attaches CatalanTagger-style isTagged for isMisspelled.
func WireCompoundRuleTagger(r *CompoundRule, tagWord func(word string) []*languagetool.AnalyzedTokenReadings) {
	if r == nil {
		return
	}
	r.TagIsTagged = func(word string) bool {
		if tagWord == nil {
			return false
		}
		readings := tagWord(word)
		if len(readings) == 0 || readings[0] == nil {
			return false
		}
		return readings[0].IsTagged()
	}
}

func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
