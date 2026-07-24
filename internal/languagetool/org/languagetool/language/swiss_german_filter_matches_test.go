package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFilterSwissGermanRuleMatches_EszettToSS(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 5, RuleID: "GERMAN_SPELLER_RULE", Suggestions: []string{"Straße", "groß"}},
	}
	out := FilterSwissGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, []string{"Strasse", "gross"}, out[0].Suggestions)
}

func TestFilterSwissGermanRuleMatches_DropAISSToSZ(t *testing.T) {
	// Java: matchingString "gross" + suggestion "groß" → drop AI orthography match.
	in := []languagetool.LocalMatch{
		{
			FromPos: 0, ToPos: 5,
			RuleID:           "AI_DE_GGEC_REPLACEMENT_ORTHOGRAPHY_SPELL",
			OriginalErrorStr: "gross",
			Suggestions:      []string{"groß"},
		},
		{
			FromPos: 6, ToPos: 10,
			RuleID:      "GERMAN_SPELLER_RULE",
			Suggestions: []string{"Haus"},
		},
	}
	out := FilterSwissGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "GERMAN_SPELLER_RULE", out[0].RuleID)
}

func TestFilterSwissGermanRuleMatches_KeepAIWithoutSurface(t *testing.T) {
	// Without OriginalErrorStr / SentenceText, do not invent drop.
	in := []languagetool.LocalMatch{
		{
			RuleID:      "AI_DE_GGEC_REPLACEMENT_ORTHOGRAPHY_SPELL",
			Suggestions: []string{"groß"},
		},
	}
	out := FilterSwissGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, []string{"gross"}, out[0].Suggestions) // still ß→ss rewrite
}

func TestFilterSwissGermanRuleMatches_DropAISSToSZ_FromSentenceText(t *testing.T) {
	// Java: matchingString = sentence.getText().substring(from,to) when originalErrorStr empty.
	in := []languagetool.LocalMatch{
		{
			FromPos: 0, ToPos: 5,
			RuleID:       "AI_DE_GGEC_REPLACEMENT_ORTHOGRAPHY_SPELL",
			SentenceText: "gross Haus",
			// sentence-local span (Java fromPos when still sentence-relative / FromPosSentence)
			FromPosSentence: 0, ToPosSentence: 5,
			Suggestions: []string{"groß"},
		},
	}
	out := FilterSwissGermanRuleMatches(in)
	require.Empty(t, out)
}

func TestFilterSwissGermanRuleMatches_CallsGermanMerge(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"foo"}},
		{FromPos: 3, ToPos: 6, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"bar"}},
	}
	out := FilterSwissGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_DE_MERGED_MATCH", out[0].RuleID)
	require.Equal(t, []string{"foobar"}, out[0].Suggestions)
}

// Java SwissGerman: both AI_DE_GGEC_REPLACEMENT_ORTHOGRAPHY_SPELL and
// AI_DE_GGEC_REPLACEMENT_ADJECTIVE_FORM are ss→ß skip IDs.
func TestFilterSwissGermanRuleMatches_DropAdjectiveFormSSToSZ(t *testing.T) {
	in := []languagetool.LocalMatch{
		{
			FromPos: 0, ToPos: 5,
			RuleID:           "AI_DE_GGEC_REPLACEMENT_ADJECTIVE_FORM",
			OriginalErrorStr: "gross",
			Suggestions:      []string{"groß"},
		},
	}
	out := FilterSwissGermanRuleMatches(in)
	require.Empty(t, out)
}

// Java: only drop when a suggestion equals surface with ss→ß; other suggestions keep the match
// and still rewrite ß→ss on remaining suggestions.
func TestFilterSwissGermanRuleMatches_KeepWhenSuggestionNotOnlySSToSZ(t *testing.T) {
	in := []languagetool.LocalMatch{
		{
			FromPos: 0, ToPos: 5,
			RuleID:           "AI_DE_GGEC_REPLACEMENT_ORTHOGRAPHY_SPELL",
			OriginalErrorStr: "gross",
			// not equal to surface.Replace("ss","ß") → do not drop
			Suggestions: []string{"grösser", "großartig"},
		},
	}
	out := FilterSwissGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, []string{"grösser", "grossartig"}, out[0].Suggestions)
}

// Non-skip AI rule IDs must not be dropped even if suggestion is ss→ß only.
func TestFilterSwissGermanRuleMatches_NoDropNonSkipAIRule(t *testing.T) {
	in := []languagetool.LocalMatch{
		{
			FromPos: 0, ToPos: 5,
			RuleID:           "AI_DE_GGEC_REPLACEMENT_NOUN",
			OriginalErrorStr: "gross",
			Suggestions:      []string{"groß"},
		},
	}
	out := FilterSwissGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, []string{"gross"}, out[0].Suggestions) // ß→ss rewrite only
}

func TestSwissGermanAdvancedTypography(t *testing.T) {
	// Swiss double quotes « »
	require.Equal(t, "Meinten Sie «entschieden» oder «entscheidend»?",
		SwissGermanAdvancedTypography(`Meinten Sie "entschieden" oder "entscheidend"?`))
	require.Equal(t, "z.\u00a0B.", SwissGermanAdvancedTypography("z.B."))
	// GermanVariant de-CH dispatches to Swiss quotes
	require.Equal(t, "Meinten Sie «entschieden»?",
		SwissGerman.ToAdvancedTypography(`Meinten Sie "entschieden"?`))
	// DE keeps German quotes
	require.Equal(t, "Meinten Sie „entschieden“?",
		GermanyGerman.ToAdvancedTypography(`Meinten Sie "entschieden"?`))
}
