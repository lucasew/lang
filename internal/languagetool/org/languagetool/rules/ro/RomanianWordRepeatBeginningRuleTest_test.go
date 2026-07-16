package ro

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRomanianWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewRomanianWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_adv":       "adv",
		"desc_repetition_beginning_word":      "word",
		"desc_repetition_beginning_thesaurus": "thes",
	})
	// Without POS tags, three same starts still fire as word repetition.
	matches := rule.MatchList(languagetool.SplitAndAnalyze("Eu merg. Eu văd. Eu vorbesc."))
	require.Equal(t, 1, len(matches))
	// two only — no error
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Eu merg. Eu văd."))))

	// Tagged adverb: build sentence manually with G-tag on first word
	g := "G"
	mk := func(text string, first string) *languagetool.AnalyzedSentence {
		// Analyze then re-tag first content token
		s := languagetool.AnalyzePlain(text)
		toks := s.GetTokensWithoutWhitespace()
		if len(toks) > 1 {
			// Replace with a reading that has adverb POS
			at := languagetool.NewAnalyzedToken(first, &g, nil)
			// can't easily replace — use HasPartial via new readings in a custom sentence
			_ = at
		}
		return s
	}
	_ = mk
	// Inject adverb tag via NewAnalyzedTokenReadings
	ss := languagetool.SentenceStartTagName
	start := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &ss, nil))
	gTag := "G0"
	adv1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Și", &gTag, nil))
	rest1a := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("merg", nil, nil))
	dot1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(".", nil, nil))
	s1 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, adv1, rest1a, dot1})

	adv2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Și", &gTag, nil))
	rest2a := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("văd", nil, nil))
	dot2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(".", nil, nil))
	s2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, adv2, rest2a, dot2})

	matches = rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.Equal(t, 1, len(matches), "two successive G-tagged Și should match as adverbs")
}
