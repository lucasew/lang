package br

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/topo.txt
var topoFS embed.FS

var (
	topoOnce sync.Once
	topoBase *rules.AbstractSimpleReplaceRule2
)

func loadTopo() *rules.AbstractSimpleReplaceRule2 {
	topoOnce.Do(func() {
		f, err := topoFS.Open("data/topo.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "BR_TOPO",
			Description:          "anvioù-lec’h e brezhoneg",
			ShortMsg:             "Anv-lec'h",
			MessageTemplate:      "Anv gallek: $match. Gwelloc'h eo ober gant $suggestions",
			SuggestionsSeparator: " pe ",
			// Java TopoReplaceRule is case-sensitive for multiword maps; use CI for surface port.
			CaseSens:     rules.CaseInsensitive,
			LanguageCode: "br",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/br/topo.txt"); err != nil {
			panic(err)
		}
		topoBase = base
	})
	return topoBase
}

// TopoReplaceRule ports org.languagetool.rules.br.TopoReplaceRule (ASR2 multiword stand-in).
type TopoReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewTopoReplaceRule(messages map[string]string) *TopoReplaceRule {
	base := loadTopo()
	r := *base
	r.Messages = messages
	return &TopoReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *TopoReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
