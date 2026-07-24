package xx

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// DemoDisambiguator2 ports
// org.languagetool.tagging.disambiguation.rules.xx.DemoDisambiguator2 —
// XmlRuleDisambiguator over Demo language rules (injected list).
type DemoDisambiguator2 struct {
	disambiguation.AbstractDisambiguator
	inner *rules.XmlRuleDisambiguator
}

func NewDemoDisambiguator2(disambigRules ...*rules.DisambiguationPatternRule) *DemoDisambiguator2 {
	return &DemoDisambiguator2{inner: rules.NewXmlRuleDisambiguator(disambigRules)}
}

func (d *DemoDisambiguator2) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if d == nil || d.inner == nil {
		return input
	}
	return d.inner.Disambiguate(input)
}

var _ disambiguation.Disambiguator = (*DemoDisambiguator2)(nil)
