package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/pt-BR/archaisms.txt
var archaismsFS embed.FS

var (
	archaismsOnce sync.Once
	archaismsBase *rules.AbstractSimpleReplaceRule2
)

func loadArchaisms() *rules.AbstractSimpleReplaceRule2 {
	archaismsOnce.Do(func() {
		f, err := archaismsFS.Open("data/pt-BR/archaisms.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "PT_ARCHAISMS_REPLACE",
			Description:          "Palavras arcaicas evitáveis",
			ShortMsg:             "Arcaísmo",
			MessageTemplate:      "\"$match\" é um arcaísmo. É preferível dizer $suggestions.",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/pt/pt-BR/archaisms.txt"); err != nil {
			panic(err)
		}
		archaismsBase = base
	})
	return archaismsBase
}

// PortugueseArchaismsRule ports org.languagetool.rules.pt.PortugueseArchaismsRule
// using the pt-BR archaisms dictionary by default.
type PortugueseArchaismsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewPortugueseArchaismsRule(messages map[string]string) *PortugueseArchaismsRule {
	base := loadArchaisms()
	r := *base
	r.Messages = messages
	return &PortugueseArchaismsRule{AbstractSimpleReplaceRule2: &r}
}

func (r *PortugueseArchaismsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
