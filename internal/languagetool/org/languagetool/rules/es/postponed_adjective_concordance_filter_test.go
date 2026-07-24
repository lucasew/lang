package es

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


// sentenceWithoutWS builds AnalyzedSentence whose GetTokensWithoutWhitespace equals toks
// (prepend SENT_START like Java non-blank stream index 0).
func sentenceWithoutWS(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptr(languagetool.SentenceStartTagName), nil))
	all := append([]*languagetool.AnalyzedTokenReadings{start}, toks...)
	return languagetool.NewAnalyzedSentence(all)
}

func TestESPostponedAdjectiveRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.es.PostponedAdjectiveConcordanceFilter"))
}

func TestPostponedAdjectiveConcordanceFilter_NoSentence(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	require.Nil(t, f.AcceptRuleMatch(nil, nil, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(&rules.RuleMatch{}, nil, 0, nil, nil))
}

// "casa bonito" — feminine noun + masculine adj → mismatch, suggestions via synth.
func TestPostponedAdjectiveConcordanceFilter_CasaBonito(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, posTagRegex string) []string {
		if posTagRegex == "A..FS.|V.P..SF|PX.FS.*" {
			return []string{"bonita"}
		}
		return nil
	}
	// tokens: [0]=SENT_START, [1]=casa NCFS000, [2]=bonito AQ0MS0
	sent := sentenceWithoutWS(
		atr("casa", "NCFS000", "casa"),
		atr("bonito", "AQ0MS0", "bonito"),
	)
	m := rules.NewRuleMatch(nil, sent, 5, 11, "msg")
	// patternTokenPos = index of adjective in non-blank tokens
	out := f.AcceptRuleMatch(m, nil, 2, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"bonita"}, out.GetSuggestedReplacements())
}

// Agreeing MS noun + MS adj → filter drops (previous agreeing noun).
func TestPostponedAdjectiveConcordanceFilter_Agreeing(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, posTagRegex string) []string {
		return []string{"should-not-appear"}
	}
	sent := sentenceWithoutWS(
		atr("libro", "NCMS000", "libro"),
		atr("bonito", "AQ0MS0", "bonito"),
	)
	m := rules.NewRuleMatch(nil, sent, 6, 12, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 2, nil, nil))
}

// No synthesizer → empty suggestions → nil (fail-closed, same as empty synth).
func TestPostponedAdjectiveConcordanceFilter_NoSynth(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	sent := sentenceWithoutWS(
		atr("casa", "NCFS000", "casa"),
		atr("bonito", "AQ0MS0", "bonito"),
	)
	m := rules.NewRuleMatch(nil, sent, 5, 11, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 2, nil, nil))
}

func TestESMatchPostagRegexp_Unknown(t *testing.T) {
	// null POS → "UNKNOWN"
	r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("x", nil, nil))
	require.True(t, esMatchPostagRegexp(r, esKEEPCOUNT)) // UNKNOWN in KEEP_COUNT
	require.False(t, esMatchPostagRegexp(r, esNOM))
}
