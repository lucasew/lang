package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicDoublePunctuationRule ports org.languagetool.rules.ar.ArabicDoublePunctuationRule.
//
// Java:
//
//	public class ArabicDoublePunctuationRule extends DoublePunctuationRule {
//	  public ArabicDoublePunctuationRule(ResourceBundle messages) {
//	    super(messages);
//	  }
//	  @Override public final String getId() { return "ARABIC_DOUBLE_PUNCTUATION"; }
//	  @Override public String getCommaCharacter() { return "،"; }
//	}
//
// Match body is inherited from DoublePunctuationRule; only id and comma character
// (Arabic comma U+060C) differ from the base rule. Java does not call setDefaultOff.
type ArabicDoublePunctuationRule struct {
	*rules.DoublePunctuationRule
}

// NewArabicDoublePunctuationRule ports ArabicDoublePunctuationRule(ResourceBundle).
func NewArabicDoublePunctuationRule(messages map[string]string) *ArabicDoublePunctuationRule {
	base := rules.NewDoublePunctuationRule(messages)
	// Java @Override getId
	base.RuleID = "ARABIC_DOUBLE_PUNCTUATION"
	// Java @Override getCommaCharacter → "،" (U+060C)
	base.CommaCharacter = "،"
	return &ArabicDoublePunctuationRule{DoublePunctuationRule: base}
}

// GetID ports ArabicDoublePunctuationRule.getId.
func (r *ArabicDoublePunctuationRule) GetID() string {
	return "ARABIC_DOUBLE_PUNCTUATION"
}

// GetCommaCharacter ports ArabicDoublePunctuationRule.getCommaCharacter.
func (r *ArabicDoublePunctuationRule) GetCommaCharacter() string {
	return "،"
}

// Match ports inherited DoublePunctuationRule.match with RuleMatch.rule = this.
func (r *ArabicDoublePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	ms := r.DoublePunctuationRule.Match(sentence)
	for _, m := range ms {
		if m != nil {
			// Java new RuleMatch(this, ...) — this is the Arabic subclass instance.
			m.Rule = r
		}
	}
	return ms
}
