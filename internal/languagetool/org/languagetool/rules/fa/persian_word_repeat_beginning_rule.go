package fa

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PersianWordRepeatBeginningRule ports org.languagetool.rules.fa.PersianWordRepeatBeginningRule.
// List of Persian adverbs for WordRepeatBeginningRule (Java @since 2.7).
type PersianWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

// persianAdverbs ports Java private static final Set<String> ADVERBS static block order.
// Surfaces are exact Java string literals (e.g. "ديگر" uses Arabic yeh U+064A).
// Multi-word members match only if the whole phrase is a single token.getToken()
// (Java ADVERBS.contains(token.getToken()); WordTokenizer splits on space).
var persianAdverbs = map[string]bool{
	"هم":          true,
	"همچنین":      true,
	"نیز":         true,
	"از یک سو":    true,
	"از یک طرف":   true,
	"از طرف ديگر": true, // Java: "ديگر" with U+064A
	"بنابراین":    true,
	"حتی":         true,
	"چنانچه":      true,
}

// NewPersianWordRepeatBeginningRule ports PersianWordRepeatBeginningRule(messages, language).
// Language is unused by this type and the core match loop (same as Greek/English WRB ports).
func NewPersianWordRepeatBeginningRule(messages map[string]string) *PersianWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	// Java @Override getId
	base.IDOverride = "PERSIAN_WORD_REPEAT_BEGINNING_RULE"
	// Java addExamplePair(Example.wrong(...), Example.fixed(...))
	base.AddExamplePair(
		rules.Wrong("همچنین، خیابان تقریباً کاملاً مسکونی است. <marker>همچنین</marker>، به افتخار یک شاعر نامگذاری شده‌است."),
		rules.Fixed("همچنین، خیابان تقریباً مسکونی است. <marker>این خیابان</marker> به افتخار یک شاعر نامگذاری شده‌است."),
	)
	r := &PersianWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	// Java @Override isAdverb only — no isException / getSuggestions overrides.
	base.IsAdverbFn = r.isAdverb
	return r
}

// isAdverb ports PersianWordRepeatBeginningRule.isAdverb:
//
//	return ADVERBS.contains(token.getToken());
//
// No punctuation stripping: Java PersianWordTokenizer splits on "،؟؛", so
// sentence-initial "همچنین،" yields token "همچنین" then "،" separately.
func (r *PersianWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	return persianAdverbs[token.GetToken()]
}
