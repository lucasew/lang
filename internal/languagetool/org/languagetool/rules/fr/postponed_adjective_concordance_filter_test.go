package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptr(s string) *string { return &s }

func atr(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(token, ptr(pos), ptr(lemma)))
}

func sentenceWithoutWS(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptr(languagetool.SentenceStartTagName), nil))
	all := append([]*languagetool.AnalyzedTokenReadings{start}, toks...)
	return languagetool.NewAnalyzedSentence(all)
}

func TestFRPostponedAdjectiveRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.fr.PostponedAdjectiveConcordanceFilter"))
}

func TestPostponedAdjectiveConcordanceFilter_NoSentence(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	require.Nil(t, f.AcceptRuleMatch(nil, nil, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(&rules.RuleMatch{}, nil, 0, nil, nil))
}

// "maison beau" — N f s + J m s → suggest "belle"
func TestPostponedAdjectiveConcordanceFilter_MaisonBeau(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, posTagRegex string) []string {
		if posTagRegex == "J [fe] sp?|V ppa f s" {
			return []string{"belle"}
		}
		return nil
	}
	sent := sentenceWithoutWS(
		atr("maison", "N f s", "maison"),
		atr("beau", "J m s", "beau"),
	)
	m := rules.NewRuleMatch(nil, sent, 7, 11, "msg")
	out := f.AcceptRuleMatch(m, nil, 2, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"belle"}, out.GetSuggestedReplacements())
}

func TestPostponedAdjectiveConcordanceFilter_Agreeing(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, posTagRegex string) []string {
		return []string{"should-not-appear"}
	}
	sent := sentenceWithoutWS(
		atr("livre", "N m s", "livre"),
		atr("beau", "J m s", "beau"),
	)
	m := rules.NewRuleMatch(nil, sent, 6, 10, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 2, nil, nil))
}

func TestPostponedAdjectiveConcordanceFilter_NoSynth(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	sent := sentenceWithoutWS(
		atr("maison", "N f s", "maison"),
		atr("beau", "J m s", "beau"),
	)
	m := rules.NewRuleMatch(nil, sent, 7, 11, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 2, nil, nil))
}

func TestFRFullMatch_UnknownNotNom(t *testing.T) {
	r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("x", nil, nil))
	require.True(t, frMatchPostagRegexp(r, frKEEPCOUNT))
	require.False(t, frMatchPostagRegexp(r, frNOM))
}
