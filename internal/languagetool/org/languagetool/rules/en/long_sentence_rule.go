package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// LongSentenceRule ports org.languagetool.rules.en.LongSentenceRule.
// Matching / max-words / ID / STYLE / picky come from core base.
type LongSentenceRule struct {
	*rules.LongSentenceRule
}

// NewLongSentenceRule ports
// LongSentenceRule(ResourceBundle messages, UserConfig userConfig, int defaultWords).
// userConfig may be nil (Java null). No invent default for defaultWords — caller
// passes it (English.getRelevantRules uses 40).
func NewLongSentenceRule(messages map[string]string, userConfig *languagetool.UserConfig, defaultWords int) *LongSentenceRule {
	// Java: super(messages, userConfig, defaultWords);
	base := rules.NewLongSentenceRuleWithUserConfig(messages, userConfig, defaultWords)
	// Java: setUrl(Tools.getUrl("https://languagetool.org/insights/post/splitting-long-sentences/"));
	base.SetURL("https://languagetool.org/insights/post/splitting-long-sentences/")
	// Java getShortMessage → "Long sentence" (also kept on base.ShortMsg for field access).
	base.ShortMsg = "Long sentence"
	return &LongSentenceRule{LongSentenceRule: base}
}

// GetShortMessage ports en.LongSentenceRule.getShortMessage().
func (r *LongSentenceRule) GetShortMessage() string {
	return "Long sentence"
}

// MatchList ports TextLevelRule.match inherited from core LongSentenceRule.
func (r *LongSentenceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.LongSentenceRule == nil {
		return nil
	}
	return r.LongSentenceRule.MatchList(sentences)
}
