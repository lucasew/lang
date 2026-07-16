package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/pt-BR/replace.txt
var brReplaceFS embed.FS

var (
	brReplaceOnce sync.Once
	brReplaceBase *rules.AbstractSimpleReplaceRule2
)

func loadBRReplace() *rules.AbstractSimpleReplaceRule2 {
	brReplaceOnce.Do(func() {
		f, err := brReplaceFS.Open("data/pt-BR/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "PT_BR_SIMPLE_REPLACE",
			Description:          "Palavras portuguesas facilmente confundidas com as do Brasil",
			ShortMsg:             "Palavra do português de Portugal",
			MessageTemplate:      "\"$match\" é uma expressão usada sobretudo em Portugal. No português brasileiro diz-se $suggestions",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   true,
			// Without NP tagger: exempt common PT given names that are proper-noun exceptions in Java.
			IsTokenException: func(atr *languagetool.AnalyzedTokenReadings) bool {
				switch atr.GetToken() {
				case "António", "Antónios", "Jerónimo", "Jerónimos",
					"antónio", "antónios", "jerónimo", "jerónimos":
					return true
				}
				return atr.IsImmunized()
			},
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/pt/pt-BR/replace.txt"); err != nil {
			panic(err)
		}
		brReplaceBase = base
	})
	return brReplaceBase
}

// BrazilianPortugueseReplaceRule ports org.languagetool.rules.pt.BrazilianPortugueseReplaceRule.
type BrazilianPortugueseReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewBrazilianPortugueseReplaceRule(messages map[string]string) *BrazilianPortugueseReplaceRule {
	base := loadBRReplace()
	r := *base
	r.Messages = messages
	return &BrazilianPortugueseReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *BrazilianPortugueseReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
