package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractSimpleReplaceRule_SkipIgnoredBySpeller(t *testing.T) {
	r := &AbstractSimpleReplaceRule{
		WrongWords: map[string][]string{"teh": {"the"}},
		ID:         "TEST_REPLACE",
	}
	// normal match
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("teh", nil, nil)),
	})
	require.NotEmpty(t, r.Match(sent))

	// ignored by speller → no match
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("teh", nil, nil))
	tok.IgnoreSpelling()
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{tok})
	require.Empty(t, r.Match(sent2))
}

func TestAbstractSimpleReplaceRule_SkipTagged(t *testing.T) {
	pos := "NN"
	r := &AbstractSimpleReplaceRule{
		WrongWords:        map[string][]string{"can": {"could"}},
		IgnoreTaggedWords: true,
		ID:                "TEST_REPLACE2",
	}
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("can", &pos, nil))
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{tok})
	require.Empty(t, r.Match(sent), "tagged word should be skipped")
}
