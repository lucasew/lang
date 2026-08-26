package fa

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PersianCommaWhitespaceRule ports org.languagetool.rules.fa.PersianCommaWhitespaceRule.
//
// Java:
//
//	public class PersianCommaWhitespaceRule extends CommaWhitespaceRule {
//	  public PersianCommaWhitespaceRule(ResourceBundle messages) {
//	    super(messages);
//	    setDefaultOff();
//	  }
//	  @Override public final String getId() { return "PERSIAN_COMMA_PARENTHESIS_WHITESPACE"; }
//	  @Override public String getCommaCharacter() { return "،"; }
//	}
//
// Match body is inherited from CommaWhitespaceRule (final in Java); only id, comma
// character (Arabic comma U+060C), and default-off differ from the base rule.
type PersianCommaWhitespaceRule struct {
	*rules.CommaWhitespaceRule
}

// NewPersianCommaWhitespaceRule ports PersianCommaWhitespaceRule(ResourceBundle).
func NewPersianCommaWhitespaceRule(messages map[string]string) *PersianCommaWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	// Java @Override getId
	base.RuleID = "PERSIAN_COMMA_PARENTHESIS_WHITESPACE"
	// Java @Override getCommaCharacter → "،" (U+060C)
	base.CommaCharacter = "،"
	// Java ctor: setDefaultOff()
	base.DefaultOff = true
	return &PersianCommaWhitespaceRule{CommaWhitespaceRule: base}
}

// GetID ports PersianCommaWhitespaceRule.getId.
func (r *PersianCommaWhitespaceRule) GetID() string {
	return "PERSIAN_COMMA_PARENTHESIS_WHITESPACE"
}

// GetCommaCharacter ports PersianCommaWhitespaceRule.getCommaCharacter.
func (r *PersianCommaWhitespaceRule) GetCommaCharacter() string {
	return "،"
}

// IsDefaultOff ports Rule.isDefaultOff after Java ctor setDefaultOff().
func (r *PersianCommaWhitespaceRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

// Match ports inherited CommaWhitespaceRule.match (final in Java) with RuleMatch.rule = this.
func (r *PersianCommaWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	ms := r.CommaWhitespaceRule.Match(sentence)
	for _, m := range ms {
		if m != nil {
			// Java new RuleMatch(this, ...) — this is the Persian subclass instance.
			m.Rule = r
		}
	}
	return ms
}
