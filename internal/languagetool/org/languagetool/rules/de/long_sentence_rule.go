package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// LongSentenceRule ports org.languagetool.rules.de.LongSentenceRule.
// Matching / STYLE / picky / minToCheckParagraph come from core base; DE overrides
// getId, getDescription, getShortMessage and adds an example pair in the ctor.
type LongSentenceRule struct {
	*rules.LongSentenceRule
}

// NewLongSentenceRule ports
// LongSentenceRule(ResourceBundle messages, UserConfig userConfig, int defaultWords).
// userConfig may be nil (Java null). Java super() resolves maxWords via
// userConfig.getConfigValueByID(getId()) with polymorphic getId() → TOO_LONG_SENTENCE_DE.
func NewLongSentenceRule(messages map[string]string, userConfig *languagetool.UserConfig, defaultWords int) *LongSentenceRule {
	// Java super: tmpMaxWords = defaultWords; if userConfig config for getId() → override.
	tmpMaxWords := defaultWords
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID("TOO_LONG_SENTENCE_DE"); len(cf) > 0 && cf[0] != nil {
			switch n := cf[0].(type) {
			case int:
				tmpMaxWords = n
			case int32:
				tmpMaxWords = int(n)
			case int64:
				tmpMaxWords = int(n)
			case float64:
				tmpMaxWords = int(n)
			}
		}
	}
	// Pass nil UserConfig to base so parent does not re-resolve with core RULE_ID.
	base := rules.NewLongSentenceRuleWithUserConfig(messages, nil, tmpMaxWords)
	// Java getId() override.
	base.RuleID = "TOO_LONG_SENTENCE_DE"
	// Java getDescription() override (not MessagesBundle long_sentence_rule_desc).
	base.Description = "Findet lange Sätze"
	// Java getShortMessage() → "Langer Satz".
	base.ShortMsg = "Langer Satz"
	// Java ctor addExamplePair(Example.wrong(...), Example.fixed(...)).
	base.AddExamplePair(
		rules.Wrong("<marker>Dies ist ein Bandwurmsatz, der immer weiter geht, obwohl das kein guter Stil ist, den man eigentlich berücksichtigen sollte, obwohl es auch andere Meinungen gibt, die aber in der Minderzahl sind, weil die meisten Autoren sich doch an die Stilvorgaben halten, wenn auch nicht alle, was aber letztendlich wiederum eine Sache des Geschmacks ist</marker>."),
		rules.Fixed("<marker>Dies ist ein kurzer Satz.</marker>"),
	)
	return &LongSentenceRule{LongSentenceRule: base}
}

// GetID ports de.LongSentenceRule.getId → "TOO_LONG_SENTENCE_DE".
func (r *LongSentenceRule) GetID() string {
	return "TOO_LONG_SENTENCE_DE"
}

// GetDescription ports de.LongSentenceRule.getDescription → "Findet lange Sätze".
func (r *LongSentenceRule) GetDescription() string {
	return "Findet lange Sätze"
}

// GetShortMessage ports de.LongSentenceRule.getShortMessage → "Langer Satz".
func (r *LongSentenceRule) GetShortMessage() string {
	return "Langer Satz"
}

// MatchList ports TextLevelRule.match inherited from core LongSentenceRule.
func (r *LongSentenceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.LongSentenceRule == nil {
		return nil
	}
	return r.LongSentenceRule.MatchList(sentences)
}
