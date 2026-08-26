package fa

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// faConjunctions ports Java PersianSpaceBeforeRule.CONJUNCTIONS:
//
//	Pattern.compile("و|به|با|تا|زیرا|چون|بنابراین|چونکه")
//
// Unanchored body; AbstractSpaceBeforeRule.Match applies Java Matcher.matches()
// full-token semantics (including shorter-left alt چون|چونکه).
var faConjunctions = regexp.MustCompile(`و|به|با|تا|زیرا|چون|بنابراین|چونکه`)

// PersianSpaceBeforeRule ports org.languagetool.rules.fa.PersianSpaceBeforeRule.
//
// Java:
//
//	public PersianSpaceBeforeRule(ResourceBundle messages, Language language) {
//	  super(messages, language);
//	  setDefaultOff();
//	}
//
// Language is accepted by Java AbstractSpaceBeforeRule but unused; Go omits it
// (same as other space-before ports).
type PersianSpaceBeforeRule struct {
	*rules.AbstractSpaceBeforeRule
}

// NewPersianSpaceBeforeRule ports the Java constructor (messages + setDefaultOff).
func NewPersianSpaceBeforeRule(messages map[string]string) *PersianSpaceBeforeRule {
	base := &rules.AbstractSpaceBeforeRule{
		ID: "FA_SPACE_BEFORE_CONJUNCTION",
		// Java PersianSpaceBeforeRule getDescription / getShort / getSuggestion
		Description:  "بررسی‌کردن فاصله قبل از حرف ربط",
		ShortMsg:     "فاصلهٔ حذف‌شده",
		Suggestion:   "فاصلهٔ قبل از حرف ربط حذف شده‌است",
		Conjunctions: faConjunctions,
		// Java setDefaultOff()
		DefaultOff: true,
	}
	// Java super: Categories.MISC.getCategory(messages)
	rules.InitSpaceBeforeMeta(base, messages)
	return &PersianSpaceBeforeRule{AbstractSpaceBeforeRule: base}
}

// Match delegates to AbstractSpaceBeforeRule.match (final in Java).
func (r *PersianSpaceBeforeRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSpaceBeforeRule.Match(sentence)
}
