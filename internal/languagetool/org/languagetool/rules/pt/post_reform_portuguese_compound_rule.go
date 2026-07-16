package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/post-reform-compounds.txt
var postReformFS embed.FS

var (
	postReformOnce sync.Once
	postReformData *rules.CompoundRuleData
)

func loadPostReformCompounds() *rules.CompoundRuleData {
	postReformOnce.Do(func() {
		f, err := postReformFS.Open("data/post-reform-compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, "/pt/post-reform-compounds.txt")
		if err != nil {
			panic(err)
		}
		postReformData = d
	})
	return postReformData
}

// PostReformPortugueseCompoundRule ports org.languagetool.rules.pt.PostReformPortugueseCompoundRule.
type PostReformPortugueseCompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewPostReformPortugueseCompoundRule(messages map[string]string) *PostReformPortugueseCompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                   messages,
		ID:                         "PT_COMPOUNDS_POST_REFORM",
		Description:                "Erro na formação da palavra composta \"$match\"",
		WithHyphenMessage:          "Esta palavra é hifenizada.",
		WithoutHyphenMessage:       "Esta palavra é composta por justaposição.",
		WithOrWithoutHyphenMessage: "Esta palavra pode ser composta por justaposição ou hifenizada.",
		ShortDesc:                  "Este conjunto forma uma palavra composta.",
		Data:                       loadPostReformCompounds(),
	}
	base.UseSubRuleSpecificIDs()
	return &PostReformPortugueseCompoundRule{AbstractCompoundRule: base}
}

func (r *PostReformPortugueseCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
