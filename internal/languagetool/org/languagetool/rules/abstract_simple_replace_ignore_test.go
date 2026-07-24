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

// Java useSubRuleSpecificIds → SpecificIdRule(toId(getId()+"_"+token), desc.replace("$match", …)).
func TestAbstractSimpleReplaceRule_SubRuleSpecificIDs(t *testing.T) {
	r := &AbstractSimpleReplaceRule{
		WrongWords:         map[string][]string{"teh": {"the"}},
		ID:                 "TEST_REPLACE",
		Description:        "Bad word: $match",
		LanguageCode:       "en",
		SubRuleSpecificIDs: true,
		MessageFn: func(tokenStr string, replacements []string) string {
			return "fix " + tokenStr
		},
	}
	sent := languagetool.AnalyzePlain("teh.")
	ms := r.Match(sent)
	require.NotEmpty(t, ms)
	idRule, ok := ms[0].GetRule().(*SpecificIdRule)
	require.True(t, ok, "must be SpecificIdRule when SubRuleSpecificIDs")
	require.NotEqual(t, "TEST_REPLACE", idRule.GetID())
	require.Contains(t, idRule.GetID(), "TEST_REPLACE")
	require.Equal(t, "Bad word: teh", idRule.GetDescription())
	// Title-case token "Teh" → suggestion still uppercased after message
	r2 := &AbstractSimpleReplaceRule{
		WrongWords:         map[string][]string{"teh": {"the"}},
		ID:                 "TEST_REPLACE",
		LanguageCode:       "en",
		SubRuleSpecificIDs: true,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage sees replacements before title-case
			require.Equal(t, []string{"the"}, replacements)
			return "msg"
		},
	}
	ms2 := r2.Match(languagetool.AnalyzePlain("Teh."))
	require.NotEmpty(t, ms2)
	require.Equal(t, []string{"The"}, ms2[0].GetSuggestedReplacements())
}
