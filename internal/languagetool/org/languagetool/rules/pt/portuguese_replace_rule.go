package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt
var ptReplaceFS embed.FS

var (
	ptReplaceOnce sync.Once
	ptReplaceMap  map[string][]string
)

func loadPTReplace() map[string][]string {
	ptReplaceOnce.Do(func() {
		f, err := ptReplaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		ptReplaceMap = m
	})
	return ptReplaceMap
}

// PortugueseReplaceRule ports org.languagetool.rules.pt.PortugueseReplaceRule.
type PortugueseReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewPortugueseReplaceRule(messages map[string]string) *PortugueseReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadPTReplace(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "PT_SIMPLE_REPLACE",
		Description:   "Palavras estrangeiras facilmente confundidas em Português",
		ShortMsg:      "Estrangeirismo",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "'" + tokenStr + "' é um estrangeirismo. Em Português é mais comum usar: " +
				joinCommaPT(replacements) + "."
		},
	}
	return &PortugueseReplaceRule{AbstractSimpleReplaceRule: base}
}

func (r *PortugueseReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
