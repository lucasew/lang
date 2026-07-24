package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func strPtrEN(s string) *string { return &s }

func atrEN(token string, startPos int, lemma string, wsBefore bool) *languagetool.AnalyzedTokenReadings {
	pos := "NN"
	at := languagetool.NewAnalyzedToken(token, &pos, strPtrEN(lemma))
	r := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{at}, startPos)
	if wsBefore {
		r.SetWhitespaceBeforeToken(" ")
	}
	return r
}

func TestEnglishConvertToSentenceCaseFilter_MeException(t *testing.T) {
	f := NewEnglishConvertToSentenceCaseFilter()
	// first token exception: normalized lower "me", first capitalizes → "Me"
	t1 := atrEN("ME", 0, "I", false)
	t2 := atrEN("AND", 3, "and", true)
	t3 := atrEN("YOU", 7, "you", true)
	m := rules.NewRuleMatch(rules.NewFakeRule("SC"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2, t3}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Me and you"}, out.GetSuggestedReplacements())

	// "me" in the middle stays lower
	t4 := atrEN("CALL", 0, "call", false)
	t5 := atrEN("ME", 5, "I", true)
	t6 := atrEN("LATER", 8, "later", true)
	m2 := rules.NewRuleMatch(rules.NewFakeRule("SC"), nil, 0, 13, "msg")
	out2 := f.AcceptRuleMatch(m2, nil, 0, []*languagetool.AnalyzedTokenReadings{t4, t5, t6}, nil)
	require.NotNil(t, out2)
	require.Equal(t, []string{"Call me later"}, out2.GetSuggestedReplacements())
}

func TestEnglishConvertToSentenceCaseFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.EnglishConvertToSentenceCaseFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.en.EnglishConvertToSentenceCaseFilter")
	require.NotNil(t, f)
}
