package fa

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PersianWordRepeatBeginningRule ports org.languagetool.rules.fa.PersianWordRepeatBeginningRule.
type PersianWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

var persianAdverbs = map[string]bool{
	"هم": true, "همچنین": true, "نیز": true,
	"از یک سو": true, "از یک طرف": true, "از طرف ديگر": true,
	"بنابراین": true, "حتی": true, "چنانچه": true,
}

func NewPersianWordRepeatBeginningRule(messages map[string]string) *PersianWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	base.IDOverride = "PERSIAN_WORD_REPEAT_BEGINNING_RULE"
	// Java: همچنین → این خیابان
	base.AddExamplePair(
		rules.Wrong("همچنین، خیابان تقریباً کاملاً مسکونی است. <marker>همچنین</marker>، به افتخار یک شاعر نامگذاری شده‌است."),
		rules.Fixed("همچنین، خیابان تقریباً مسکونی است. <marker>این خیابان</marker> به افتخار یک شاعر نامگذاری شده‌است."),
	)
	r := &PersianWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	base.IsAdverbFn = r.isAdverb
	return r
}

func (r *PersianWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	// WordTokenizer does not split Arabic comma U+060C, so "همچنین،" may stay attached.
	return persianAdverbs[stripPersianPunct(token.GetToken())]
}

func stripPersianPunct(s string) string {
	return strings.TrimRight(s, "،,.;:!?؟؛")
}
