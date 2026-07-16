package ro

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RomanianWordRepeatBeginningRule ports org.languagetool.rules.ro.RomanianWordRepeatBeginningRule.
// Adverb = POS tag starting with "G" (see resource/ro/coduri.html).
type RomanianWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
	allowAmbiguousAdverbs bool
}

func NewRomanianWordRepeatBeginningRule(messages map[string]string) *RomanianWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	base.IDOverride = "ROMANIAN_WORD_REPEAT_BEGINNING_RULE"
	r := &RomanianWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	base.IsAdverbFn = r.isAdverb
	return r
}

func (r *RomanianWordRepeatBeginningRule) AllowAmbiguousAdverbs() bool {
	return r.allowAmbiguousAdverbs
}

func (r *RomanianWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	isAdverb := false
	for _, analyzedToken := range token.GetReadings() {
		if analyzedToken == nil {
			continue
		}
		tag := analyzedToken.GetPOSTag()
		if tag == nil {
			continue
		}
		if strings.HasPrefix(*tag, "G") {
			isAdverb = true
		} else if !r.allowAmbiguousAdverbs {
			return false
		}
	}
	return isAdverb
}
