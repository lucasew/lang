package fa

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PersianDoublePunctuationRule ports org.languagetool.rules.fa.PersianDoublePunctuationRule.
//
// Java:
//
//	public class PersianDoublePunctuationRule extends DoublePunctuationRule {
//	  public PersianDoublePunctuationRule(ResourceBundle messages) {
//	    super(messages);
//	  }
//	  @Override public final String getId() { return "PERSIAN_DOUBLE_PUNCTUATION"; }
//	  @Override public String getCommaCharacter() { return "،"; }
//	}
//
// Match body is inherited from DoublePunctuationRule; only id and comma character
// (Arabic comma U+060C) differ from the base rule. Java does not call setDefaultOff.
type PersianDoublePunctuationRule struct {
	*rules.DoublePunctuationRule
}

// NewPersianDoublePunctuationRule ports PersianDoublePunctuationRule(ResourceBundle).
func NewPersianDoublePunctuationRule(messages map[string]string) *PersianDoublePunctuationRule {
	base := rules.NewDoublePunctuationRule(messages)
	// Java @Override getId
	base.RuleID = "PERSIAN_DOUBLE_PUNCTUATION"
	// Java @Override getCommaCharacter → "،" (U+060C)
	base.CommaCharacter = "،"
	return &PersianDoublePunctuationRule{DoublePunctuationRule: base}
}

// GetID ports PersianDoublePunctuationRule.getId.
func (r *PersianDoublePunctuationRule) GetID() string {
	return "PERSIAN_DOUBLE_PUNCTUATION"
}

// GetCommaCharacter ports PersianDoublePunctuationRule.getCommaCharacter.
func (r *PersianDoublePunctuationRule) GetCommaCharacter() string {
	return "،"
}

// Match ports inherited DoublePunctuationRule.match with RuleMatch.rule = this.
func (r *PersianDoublePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	ms := r.DoublePunctuationRule.Match(sentence)
	for _, m := range ms {
		if m != nil {
			// Java new RuleMatch(this, ...) — this is the Persian subclass instance.
			m.Rule = r
		}
	}
	return ms
}
