package es

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SpanishWordRepeatRule ports org.languagetool.rules.es.SpanishWordRepeatRule.
type SpanishWordRepeatRule struct {
	*rules.WordRepeatRule
}

func NewSpanishWordRepeatRule(messages map[string]string) *SpanishWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "SPANISH_WORD_REPEAT_RULE"
	r := &SpanishWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.esIgnore
	return r
}

func (r *SpanishWordRepeatRule) esIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position > 0 && (tokens[position].HasPosTag("_allow_repeat") || tokens[position-1].HasPosTag("_allow_repeat")) {
		return true
	}
	// Surface stand-ins without Spanish tagger for common false alarms in tests.
	if position > 0 {
		prev, cur := tokens[position-1].GetToken(), tokens[position].GetToken()
		// "Bienvenido/a a" → a after slash
		if strings.EqualFold(prev, "a") && strings.EqualFold(cur, "a") && position >= 2 {
			if tokens[position-2].GetToken() == "/" {
				return true
			}
		}
		// "HUCHA-GANGA.ES es" — token before "es" is often "ES" after domain split,
		// but only when the previous token is uppercase TLD-like, not lowercase "es".
		if prev == "ES" && strings.EqualFold(cur, "es") {
			return true
		}
		if strings.HasSuffix(prev, ".ES") && strings.EqualFold(cur, "es") {
			return true
		}
	}
	return false
}
