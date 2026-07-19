package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/diacritics.txt
var diacriticsFS embed.FS

var (
	diacriticsOnce sync.Once
	diacriticsBase *rules.AbstractSimpleReplaceRule2
)

func loadDiacritics() *rules.AbstractSimpleReplaceRule2 {
	diacriticsOnce.Do(func() {
		f, err := diacriticsFS.Open("data/diacritics.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "PT_DIACRITICS_REPLACE",
			Description:          "Palavras estrangeiras com diacríticos: $match",
			ShortMsg:             "A palavra estrangeira original tem diacrítico",
			MessageTemplate:      "'$match' é uma expressão estrangeira importada cuja grafia tem diacríticos. É preferível escrever $suggestions",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/pt/diacritics.txt"); err != nil {
			panic(err)
		}
		// Java: coupe → coupé
		base.AddExamplePair(
			rules.Wrong("<marker>coupe</marker>"),
			rules.Fixed("<marker>coupé</marker>"),
		)
		diacriticsBase = base
	})
	return diacriticsBase
}

// PortugueseDiacriticsRule ports org.languagetool.rules.pt.PortugueseDiacriticsRule.
// Note: Java PortugueseDiacriticsRuleTest targets grammar rulegroup DIACRITICS (dialect
// bebé/bebê), not this ASR2 rule. Unit tests exercise the dictionary file directly.
type PortugueseDiacriticsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewPortugueseDiacriticsRule(messages map[string]string) *PortugueseDiacriticsRule {
	base := loadDiacritics()
	r := *base
	r.Messages = messages
	return &PortugueseDiacriticsRule{AbstractSimpleReplaceRule2: &r}
}

func (r *PortugueseDiacriticsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
