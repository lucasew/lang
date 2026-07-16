package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/pt-PT/replace.txt
var ptPTReplaceFS embed.FS

var (
	ptPTReplaceOnce sync.Once
	ptPTReplaceBase *rules.AbstractSimpleReplaceRule2
)

func loadPTPTReplace() *rules.AbstractSimpleReplaceRule2 {
	ptPTReplaceOnce.Do(func() {
		f, err := ptPTReplaceFS.Open("data/pt-PT/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "PT_PT_SIMPLE_REPLACE",
			Description:          "Brasileirismo: 1. palavras confundidas com as de Portugal",
			ShortMsg:             "Palavra de português do Brasil",
			MessageTemplate:      "'$match' é uma expressão brasileira, em português de Portugal utiliza-se: $suggestions",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/pt/pt-PT/replace.txt"); err != nil {
			panic(err)
		}
		ptPTReplaceBase = base
	})
	return ptPTReplaceBase
}

// PortugalPortugueseReplaceRule ports org.languagetool.rules.pt.PortugalPortugueseReplaceRule.
type PortugalPortugueseReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewPortugalPortugueseReplaceRule(messages map[string]string) *PortugalPortugueseReplaceRule {
	base := loadPTPTReplace()
	r := *base
	r.Messages = messages
	return &PortugalPortugueseReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *PortugalPortugueseReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
