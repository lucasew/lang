package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// DemoRule ports org.languagetool.rules.DemoRule — flags the token "demo".
type DemoRule struct{}

func NewDemoRule() *DemoRule { return &DemoRule{} }

func (r *DemoRule) GetID() string          { return "DEMO_RULE" }
func (r *DemoRule) GetDescription() string { return "A demo rule that just prints the text analysis" }

// Match flags each non-whitespace token equal to "demo".
func (r *DemoRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	if sentence == nil {
		return nil
	}
	var out []*RuleMatch
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		if tok.GetToken() == "demo" {
			m := NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), "The demo rule thinks this looks wrong")
			m.SetSuggestedReplacement("blablah")
			out = append(out, m)
		}
	}
	return out
}
