package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_operationnames.txt
var opNamesFS embed.FS

var (
	opNamesOnce sync.Once
	opNamesMap  map[string][]string
)

func loadOperationNames() map[string][]string {
	opNamesOnce.Do(func() {
		f, err := opNamesFS.Open("data/replace_operationnames.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		opNamesMap = m
	})
	return opNamesMap
}

// ReplaceOperationNamesRule ports org.languagetool.rules.ca.ReplaceOperationNamesRule
// (surface dictionary path; POS/synthesizer refinements not modeled).
type ReplaceOperationNamesRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewReplaceOperationNamesRule(messages map[string]string) *ReplaceOperationNamesRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadOperationNames(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "NOMS_OPERACIONS",
		Description:   "S'ha d'evitar com a nom d'operació tècnica: $match",
		ShortMsg:      "Forma preferible",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Si és el nom d'una operació tècnica, val més usar una altra forma."
		},
	}
	return &ReplaceOperationNamesRule{AbstractSimpleReplaceRule: base}
}

func (r *ReplaceOperationNamesRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
