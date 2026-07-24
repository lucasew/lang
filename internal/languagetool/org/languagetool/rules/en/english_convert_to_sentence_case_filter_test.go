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

func TestEnglishTokenIsException(t *testing.T) {
	// Java: return s.equals("me"); exact equals on already-lowercased form from core.
	require.True(t, englishTokenIsException("me"))
	require.False(t, englishTokenIsException("Me"))
	require.False(t, englishTokenIsException("ME"))
	require.False(t, englishTokenIsException("meh"))
	require.False(t, englishTokenIsException(""))
	require.False(t, englishTokenIsException("i"))
}

func TestEnglishConvertToSentenceCaseFilter_MeException(t *testing.T) {
	f := NewEnglishConvertToSentenceCaseFilter()
	// First token: core lowercases "ME" → "me", exception returns "me", then
	// firstDone applies UppercaseFirstChar → "Me"; rest lower lemmas stay lower.
	t1 := atrEN("ME", 0, "I", false)
	t2 := atrEN("AND", 3, "and", true)
	t3 := atrEN("YOU", 7, "you", true)
	m := rules.NewRuleMatch(rules.NewFakeRule("SC"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2, t3}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Me and you"}, out.GetSuggestedReplacements())

	// Mid-token "ME": exception → keep lower "me" (without EN exception, lemma "I"
	// is capitalized → "Me").
	t4 := atrEN("CALL", 0, "call", false)
	t5 := atrEN("ME", 5, "I", true)
	t6 := atrEN("LATER", 8, "later", true)
	m2 := rules.NewRuleMatch(rules.NewFakeRule("SC"), nil, 0, 13, "msg")
	out2 := f.AcceptRuleMatch(m2, nil, 0, []*languagetool.AnalyzedTokenReadings{t4, t5, t6}, nil)
	require.NotNil(t, out2)
	require.Equal(t, []string{"Call me later"}, out2.GetSuggestedReplacements())

	// Core without EN exception: mid "ME"/lemma "I" → capitalized "Me".
	core := rules.NewConvertToSentenceCaseFilter()
	t4b := atrEN("CALL", 0, "call", false)
	t5b := atrEN("ME", 5, "I", true)
	t6b := atrEN("LATER", 8, "later", true)
	m3 := rules.NewRuleMatch(rules.NewFakeRule("SC"), nil, 0, 13, "msg")
	outCore := core.AcceptRuleMatch(m3, nil, 0, []*languagetool.AnalyzedTokenReadings{t4b, t5b, t6b}, nil)
	require.NotNil(t, outCore)
	require.Equal(t, []string{"Call Me later"}, outCore.GetSuggestedReplacements())
}

func TestEnglishConvertToSentenceCaseFilter_NonExceptionSentenceCase(t *testing.T) {
	// Non-exception tokens still sentence-case via core ConvertToSentenceCaseFilter.
	f := NewEnglishConvertToSentenceCaseFilter()
	t1 := atrEN("HELLO", 0, "hello", false)
	t2 := atrEN("WORLD", 6, "world", true)
	m := rules.NewRuleMatch(rules.NewFakeRule("SC"), nil, 0, 11, "msg")
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Hello world"}, out.GetSuggestedReplacements())

	// Mid-token that is not "me" is not forced lower by EN exception.
	t3 := atrEN("CALL", 0, "call", false)
	t4 := atrEN("HIM", 5, "him", true)
	t5 := atrEN("LATER", 9, "later", true)
	m2 := rules.NewRuleMatch(rules.NewFakeRule("SC"), nil, 0, 14, "msg")
	out2 := f.AcceptRuleMatch(m2, nil, 0, []*languagetool.AnalyzedTokenReadings{t3, t4, t5}, nil)
	require.NotNil(t, out2)
	require.Equal(t, []string{"Call him later"}, out2.GetSuggestedReplacements())
}

func TestEnglishConvertToSentenceCaseFilter_Registered(t *testing.T) {
	const className = "org.languagetool.rules.en.EnglishConvertToSentenceCaseFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(className))
	f := patterns.GlobalRuleFilterCreator.GetFilter(className)
	require.NotNil(t, f)

	// Registered instance applies the "me" exception (same as NewEnglish…).
	t1 := atrEN("CALL", 0, "call", false)
	t2 := atrEN("ME", 5, "I", true)
	t3 := atrEN("LATER", 8, "later", true)
	m := rules.NewRuleMatch(rules.NewFakeRule("SC"), nil, 0, 13, "msg")
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2, t3}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Call me later"}, out.GetSuggestedReplacements())
}
