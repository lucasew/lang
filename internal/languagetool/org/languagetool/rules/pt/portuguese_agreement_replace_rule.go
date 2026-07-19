package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/AOreplace.txt
var aoReplaceFS embed.FS

var (
	aoReplaceOnce sync.Once
	aoReplaceMap  map[string][]string
)

func loadAOReplace() map[string][]string {
	aoReplaceOnce.Do(func() {
		f, err := aoReplaceFS.Open("data/AOreplace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		aoReplaceMap = m
	})
	return aoReplaceMap
}

// PortugueseAgreementReplaceRule ports org.languagetool.rules.pt.PortugueseAgreementReplaceRule.
type PortugueseAgreementReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewPortugueseAgreementReplaceRule(messages map[string]string) *PortugueseAgreementReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadAOReplace(),
		CaseSensitive: false,
		CheckLemmas:   true, // Java default checkLemmas true
		ID:            "PT_AGREEMENT_REPLACE",
		Description:   "Palavras alteradas pelo Acordo Ortográfico de 90",
		ShortMsg:      "Forma do Acordo Ortográfico de 45.",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "'" + tokenStr + "' é uma forma do antigo acordo ortográfico. No novo acordo ortográfico, a palavra escreve-se assim: " +
				joinCommaPT(replacements) + "."
		},
	}
	return &PortugueseAgreementReplaceRule{AbstractSimpleReplaceRule: base}
}

func joinCommaPT(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

func (r *PortugueseAgreementReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
