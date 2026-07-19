package fa

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PersianWordRepeatRule ports org.languagetool.rules.fa.PersianWordRepeatRule.
type PersianWordRepeatRule struct {
	*rules.WordRepeatRule
}

func NewPersianWordRepeatRule(messages map[string]string) *PersianWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "PERSIAN_WORD_REPEAT_RULE"
	// Java: برای برای → برای
	base.AddExamplePair(
		rules.Wrong("این کار <marker>برای برای</marker> تو بود."),
		rules.Fixed("این کار <marker>برای</marker> تو بود."),
	)
	r := &PersianWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.faIgnore
	return r
}

func (r *PersianWordRepeatRule) faIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position <= 0 {
		return false
	}
	for _, w := range []string{"لی", "سی", "لک", "ریز", "جز", "کل"} {
		if strings.EqualFold(tokens[position-1].GetToken(), w) && strings.EqualFold(tokens[position].GetToken(), w) {
			return true
		}
	}
	return false
}
