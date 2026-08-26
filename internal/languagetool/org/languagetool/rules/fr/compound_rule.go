package fr

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
		d, err := rules.NewCompoundRuleData(f, "/fr/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// CompoundRule ports org.languagetool.rules.fr.CompoundRule.
// isMisspelled uses TagIsTagged (Java FrenchTagger.INSTANCE.tag(...).isTagged()).
// Without TagIsTagged, isMisspelled is false (AbstractCompoundRule default — keep suggestions).
type CompoundRule struct {
	*rules.AbstractCompoundRule
	// TagIsTagged ports FrenchTagger.tag([word])[0].isTagged(); when nil, misspelled=false.
	TagIsTagged func(word string) bool
}

// NewCompoundRule constructs FR_COMPOUNDS.
func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:                          "FR_COMPOUNDS",
		Description:                 "Mots avec trait d’union : $match",
		WithHyphenMessage:           "Écrivez avec un trait d’union.",
		WithoutHyphenMessage:        "Écrivez avec un mot seul sans espace ni trait d’union.",
		WithOrWithoutHyphenMessage:  "Écrivez avec un mot seul ou avec trait d’union.",
		ShortDesc:                   "Erreur de trait d’union",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	base.UseSubRuleSpecificIDs()
	rules.InitCompoundRuleMeta(base, messages)
	// Java: Haut Rhin → Haut-Rhin
	base.AddExamplePair(
		rules.Wrong("Le <marker>Haut Rhin</marker>."),
		rules.Fixed("Le <marker>Haut-Rhin</marker>."),
	)
	r := &CompoundRule{AbstractCompoundRule: base}
	// Java CompoundRule.isMisspelled: !FrenchTagger.INSTANCE.tag(...).isTagged()
	base.IsMisspelled = func(word string) bool {
		if r.TagIsTagged == nil {
			return false
		}
		return !r.TagIsTagged(word)
	}
	return r
}

// WireCompoundRuleTagger attaches FrenchTagger-style isTagged checks
// (Java FrenchTagger.INSTANCE used by CompoundRule.isMisspelled).
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

// Match delegates to AbstractCompoundRule.
func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
