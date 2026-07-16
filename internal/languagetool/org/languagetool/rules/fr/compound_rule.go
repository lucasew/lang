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
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

// NewCompoundRule constructs FR_COMPOUNDS.
func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                    messages,
		ID:                          "FR_COMPOUNDS",
		Description:                 "Mots avec trait d’union : $match",
		WithHyphenMessage:           "Écrivez avec un trait d’union.",
		WithoutHyphenMessage:        "Écrivez avec un mot seul sans espace ni trait d’union.",
		WithOrWithoutHyphenMessage:  "Écrivez avec un mot seul ou avec trait d’union.",
		ShortDesc:                   "Erreur de trait d’union",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
		// Without FrenchTagger: treat suggestions as correctly spelled (Java isMisspelled default false).
	}
	base.UseSubRuleSpecificIDs()
	return &CompoundRule{AbstractCompoundRule: base}
}

// Match delegates to AbstractCompoundRule.
func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
