package km

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// kmConjunctions ports Java KhmerSpaceBeforeRule.CONJUNCTIONS:
//
//	Pattern.compile("ដើម្បី|និង|ពីព្រោះ", Pattern.CASE_INSENSITIVE|Pattern.UNICODE_CASE)
//
// Unanchored body; AbstractSpaceBeforeRule.Match applies Java Matcher.matches()
// full-token semantics via javaMatchesAnchored. CASE_INSENSITIVE|UNICODE_CASE is
// a no-op on these Khmer-script conjunctions (no case folding); RE2 has no
// UNICODE_CASE flag twin, same as other space-before ports.
var kmConjunctions = regexp.MustCompile(`ដើម្បី|និង|ពីព្រោះ`)

// KhmerSpaceBeforeRule ports org.languagetool.rules.km.KhmerSpaceBeforeRule.
//
// Java:
//
//	public KhmerSpaceBeforeRule(ResourceBundle messages, Language language) {
//	  super(messages, language);
//	}
//
// Does not override getDescription/getShort/getSuggestion (Abstract English defaults).
// Does not call setDefaultOff (unlike PersianSpaceBeforeRule).
// Language is accepted by Java AbstractSpaceBeforeRule but unused; Go omits it.
type KhmerSpaceBeforeRule struct {
	*rules.AbstractSpaceBeforeRule
}

// NewKhmerSpaceBeforeRule ports the Java constructor (messages only; language unused).
func NewKhmerSpaceBeforeRule(messages map[string]string) *KhmerSpaceBeforeRule {
	base := &rules.AbstractSpaceBeforeRule{
		ID: "KM_SPACE_BEFORE_CONJUNCTION",
		// Java inherits AbstractSpaceBeforeRule English defaults (no KM override).
		Description:  "Checks for missing space before some conjunctions",
		ShortMsg:     "Missing white space",
		Suggestion:   "Missing white space before conjunction",
		Conjunctions: kmConjunctions,
		// Java does not call setDefaultOff()
	}
	// Java super: Categories.MISC.getCategory(messages)
	rules.InitSpaceBeforeMeta(base, messages)
	return &KhmerSpaceBeforeRule{AbstractSpaceBeforeRule: base}
}

// Match delegates to AbstractSpaceBeforeRule.match (final in Java).
func (r *KhmerSpaceBeforeRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSpaceBeforeRule.Match(sentence)
}
