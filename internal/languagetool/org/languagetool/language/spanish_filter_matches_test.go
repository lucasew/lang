package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestFilterSpanishRuleMatches_DropObsoleteDiacritic(t *testing.T) {
	// Java suggestionsToAvoid — drop AI_ES_GGEC with single obsolete-diacritic suggestion.
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 4, RuleID: "AI_ES_GGEC_X", Suggestions: []string{"sólo"}},
		{FromPos: 5, ToPos: 8, RuleID: "OTHER"},
	}
	out := FilterSpanishRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "OTHER", out[0].RuleID)
}

func TestFilterSpanishRuleMatches_DropTrailingPeriod(t *testing.T) {
	in := []languagetool.LocalMatch{
		{
			RuleID: "AI_ES_GGEC_MISSING_PUNCTUATION", Suggestions: []string{"Hola."},
			SentenceText: "Hola  ",
		},
		{RuleID: "ES_OTHER"},
	}
	out := FilterSpanishRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "ES_OTHER", out[0].RuleID)
}

func TestFilterSpanishRuleMatches_KeepPeriodWithoutSentenceText(t *testing.T) {
	in := []languagetool.LocalMatch{
		{RuleID: "AI_ES_GGEC_MISSING_PUNCTUATION", Suggestions: []string{"Hola."}},
	}
	out := FilterSpanishRuleMatches(in)
	require.Len(t, out, 1) // fail-closed
}

func TestFilterSpanishRuleMatches_DropLowercaseSentenceStart(t *testing.T) {
	// sentence.trim().startsWith(uppercaseFirstChar(suggestion))
	in := []languagetool.LocalMatch{
		{
			RuleID: "AI_ES_GGEC_X", Suggestions: []string{"hola"},
			SentenceText: "  Hola mundo",
		},
	}
	out := FilterSpanishRuleMatches(in)
	require.Empty(t, out)
}

func TestFilterSpanishRuleMatches_CasingRewrite(t *testing.T) {
	in := []languagetool.LocalMatch{
		{
			RuleID: "AI_ES_GGEC_REPLACEMENT_ORTHOGRAPHY_X", Suggestions: []string{"Madrid"},
			OriginalErrorStr: "madrid",
		},
	}
	out := FilterSpanishRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_ES_GGEC_REPLACEMENT_CASING_X", out[0].RuleID)
	require.Equal(t, "typographical", out[0].IssueType)
	require.Equal(t, "CASING", out[0].CategoryID)
	require.Equal(t, "Mayúsculas y minúsculas", out[0].ShortMessage)
	require.Equal(t, "Mayúsculas y minúsculas recomendadas.", out[0].Message)
}

func TestFilterSpanishRuleMatches_VoseoHook(t *testing.T) {
	prev := SpanishSuggestionIsVoseo
	t.Cleanup(func() { SpanishSuggestionIsVoseo = prev })
	SpanishSuggestionIsVoseo = func(s string) bool { return s == "tenés" }
	in := []languagetool.LocalMatch{
		{RuleID: "AI_ES_GGEC_X", Suggestions: []string{"tenés"}},
		{RuleID: "AI_ES_GGEC_Y", Suggestions: []string{"tienes"}},
	}
	out := FilterSpanishRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_ES_GGEC_Y", out[0].RuleID)
}

func TestFilterSpanishRuleMatches_VoseoWithoutDictKeeps(t *testing.T) {
	// Fail-closed: empty WordTagger → no V....V.* POS → keep (not invent drop).
	prevWT := SpanishVoseoWordTagger
	prevFn := SpanishSuggestionIsVoseo
	t.Cleanup(func() {
		SpanishVoseoWordTagger = prevWT
		SpanishSuggestionIsVoseo = prevFn
	})
	SpanishVoseoWordTagger = nil
	SpanishSuggestionIsVoseo = SpanishSuggestionIsVoseoDefault
	in := []languagetool.LocalMatch{
		{RuleID: "AI_ES_GGEC_X", Suggestions: []string{"tenés"}},
	}
	out := FilterSpanishRuleMatches(in)
	require.Len(t, out, 1)
}

func TestFilterSpanishRuleMatches_VoseoPosTagDrop(t *testing.T) {
	// Java: tagger.tag(suggestion).matchesPosTagRegex("V....V.*")
	prevWT := SpanishVoseoWordTagger
	prevFn := SpanishSuggestionIsVoseo
	t.Cleanup(func() {
		SpanishVoseoWordTagger = prevWT
		SpanishSuggestionIsVoseo = prevFn
	})
	SpanishVoseoWordTagger = tagging.MapWordTagger{
		// Java Pattern "V....V.*" (full match): 6th char must be V.
		// Synthetic POS tags that exercise the exact regex (dict may use longer forms).
		"tenés":  {tagging.NewTaggedWord("tener", "VMIP2V0")},
		"tienes": {tagging.NewTaggedWord("tener", "VMIP2S0")},
	}
	SpanishSuggestionIsVoseo = SpanishSuggestionIsVoseoDefault
	in := []languagetool.LocalMatch{
		{RuleID: "AI_ES_GGEC_X", Suggestions: []string{"tenés"}},
		{RuleID: "AI_ES_GGEC_Y", Suggestions: []string{"tienes"}},
	}
	out := FilterSpanishRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_ES_GGEC_Y", out[0].RuleID)
}

func TestFilterSpanishRuleMatches_NonGGECPassthrough(t *testing.T) {
	in := []languagetool.LocalMatch{
		{RuleID: "ES_AGREEMENT", Suggestions: []string{"sólo"}},
	}
	out := FilterSpanishRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "ES_AGREEMENT", out[0].RuleID)
}
