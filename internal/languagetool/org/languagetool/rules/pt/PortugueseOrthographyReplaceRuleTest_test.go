package pt

// Twin of languagetool-language-modules/pt/.../PortugueseOrthographyReplaceRuleTest.java
//
// Analysis path:
// - Plain cases use AnalyzePlain (WordTokenizer + SENT_START). Token surfaces for
//   "Já/Ja", "você/voce" match without POS/lemmas for this rule’s dictionary path.
// - Java assertNoMatches("Disse-me sotto voce.") relies on JLanguageTool(Portuguese)
//   analysis: PortugueseHybridDisambiguator MultiWordChunker on official
//   /pt/multiwords.txt ("sotto voce") with setIgnoreSpelling(true). AbstractSimpleReplaceRule
//   skips IsIgnoredBySpeller tokens. We run the same MultiWordChunker stage on AnalyzePlain
//   output (not manual Immunize invent). Without that stage, "voce" still matches (fail closed).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	disambpt "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/pt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Twin of PortugueseOrthographyReplaceRuleTest.testRule
func TestPortugueseOrthographyReplaceRule_Rule(t *testing.T) {
	// Java: rule = new PortugueseOrthographyReplaceRule(TestTools.getMessages("pt"), lt.getLanguage())
	//       lt = new JLanguageTool(Portuguese.getInstance())
	rule := NewPortugueseOrthographyReplaceRule(nil)

	// Java assertNoMatches / assertSingleMatch
	assertPortugueseOrthographyNoMatches(t, rule, "Já volto.")
	assertPortugueseOrthographySingleMatch(t, rule, "Ja volto.", "Já")

	assertPortugueseOrthographyNoMatches(t, rule, "Gosto de você.")
	assertPortugueseOrthographySingleMatch(t, rule, "Gosto de voce.", "você")

	// Java: multi-token spelling of Italian expression via multiwords + ignore_spelling
	assertPortugueseOrthographyNoMatches(t, rule, analyzePTWithMultiwords(t, "Disse-me sotto voce."))
}

// Fail-closed: without MultiWordChunker, "voce" is a replace_orthography.txt key.
func TestPortugueseOrthographyReplaceRule_FailClosedWithoutMultiword(t *testing.T) {
	rule := NewPortugueseOrthographyReplaceRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Disse-me sotto voce."))
	require.NotEmpty(t, matches, "without multiword ignore_spelling, voce must still match (no invent skip)")
	require.Equal(t, "você", matches[0].GetSuggestedReplacements()[0])
}

// assertPortugueseOrthographyNoMatches ports assertNoMatches(String).
func assertPortugueseOrthographyNoMatches(t *testing.T, rule *PortugueseOrthographyReplaceRule, sentence any) {
	t.Helper()
	var sent *languagetool.AnalyzedSentence
	switch v := sentence.(type) {
	case string:
		sent = languagetool.AnalyzePlain(v)
	case *languagetool.AnalyzedSentence:
		sent = v
	default:
		t.Fatalf("unsupported sentence type %T", sentence)
	}
	matches := rule.Match(sent)
	require.Equal(t, 0, len(matches), "expected 0 matches for sentence: %v", sentence)
}

// assertPortugueseOrthographySingleMatch ports assertSingleMatch(String, String...).
func assertPortugueseOrthographySingleMatch(t *testing.T, rule *PortugueseOrthographyReplaceRule, sentence string, suggestions ...string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "sentence %q", sentence)
	// Java assertRuleId: match.getRule().getId().startsWith("PT_SIMPLE_REPLACE_ORTHOGRAPHY")
	assertPortugueseOrthographyRuleID(t, matches[0])
	got := matches[0].GetSuggestedReplacements()
	require.Equal(t, len(suggestions), len(got), "sentence %q suggestions %v", sentence, got)
	for i, s := range suggestions {
		require.Equal(t, s, got[i], "sentence %q suggestion[%d]", sentence, i)
	}
}

// assertPortugueseOrthographyRuleID ports assertRuleId(RuleMatch).
func assertPortugueseOrthographyRuleID(t *testing.T, match *rules.RuleMatch) {
	t.Helper()
	// useSubRuleSpecificIds → SpecificIdRule whose id starts with base id
	sr, ok := match.GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected (useSubRuleSpecificIds)")
	require.True(t, strings.HasPrefix(sr.GetID(), PortugueseSimpleReplaceOrthographyID),
		"rule id %q should start with %s", sr.GetID(), PortugueseSimpleReplaceOrthographyID)
}

// analyzePTWithMultiwords runs official /pt/multiwords.txt MultiWordChunker
// (PortugueseHybridDisambiguator.chunker stage: ignore_spelling on multiword spans).
func analyzePTWithMultiwords(t *testing.T, text string) *languagetool.AnalyzedSentence {
	t.Helper()
	plain := languagetool.AnalyzePlain(text)
	c := disambpt.PortugueseMultiWordChunker()
	if c == nil {
		t.Skip("official pt/multiwords.txt not discoverable (Java PortugueseHybridDisambiguator.chunker)")
	}
	return c.Disambiguate(plain)
}

// Field twin of PortugueseOrthographyReplaceRule constructor + overrides.
func TestPortugueseOrthographyReplaceRule_Metadata(t *testing.T) {
	// nil messages → MessagesBundle.properties English fallback (TestTools.getMessages("pt"))
	rule := NewPortugueseOrthographyReplaceRule(nil)

	require.Equal(t, PortugueseSimpleReplaceOrthographyID, rule.GetID())
	require.Equal(t, "PT_SIMPLE_REPLACE_ORTHOGRAPHY", rule.GetID())
	// messages.getString("desc_spelling")
	require.Equal(t, orthographyDefaultDescSpelling, rule.GetDescription())
	require.Equal(t, "Possible spelling mistake", rule.GetDescription())
	// messages.getString("desc_spelling_short")
	require.Equal(t, orthographyDefaultDescSpellingShort, rule.GetShort())
	require.Equal(t, "Spelling mistake", rule.GetShort())
	// messages.getString("spelling")
	require.Equal(t, orthographyDefaultSpelling, rule.GetMessage())
	require.Equal(t, "Possible spelling mistake found.", rule.GetMessage())

	require.Equal(t, "pt", rule.GetLocaleCode())
	require.False(t, rule.CaseSensitive)
	require.True(t, rule.CheckLemmas)
	require.True(t, rule.SubRuleSpecificIDs)

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Official dictionary loaded (same keys as replace_orthography.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"já"}, ww["ja"])
	require.Equal(t, []string{"você"}, ww["voce"])

	// Message keys honor supplied ResourceBundle map (e.g. pt_PT surfaces)
	ptMsgs := map[string]string{
		"desc_spelling":       "Possível erro ortográfico",
		"desc_spelling_short": "Erro ortográfico",
		"spelling":            "Possível erro ortográfico",
		"category_typo":       "Erros ortográficos",
	}
	rulePT := NewPortugueseOrthographyReplaceRule(ptMsgs)
	require.Equal(t, "Possível erro ortográfico", rulePT.GetDescription())
	require.Equal(t, "Erro ortográfico", rulePT.GetShort())
	require.Equal(t, "Possível erro ortográfico", rulePT.GetMessage())
	require.Equal(t, "Erros ortográficos", rulePT.GetCategory().GetName())

	// Sub-rule specific id + match message/short (createRuleMatch)
	matches := rule.Match(languagetool.AnalyzePlain("Ja volto."))
	require.Len(t, matches, 1)
	require.Equal(t, orthographyDefaultSpelling, matches[0].GetMessage())
	require.Equal(t, orthographyDefaultDescSpellingShort, matches[0].GetShortMessage())
	wantID := tools.ToId(PortugueseSimpleReplaceOrthographyID+"_"+"Ja", "pt")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok)
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_SIMPLE_REPLACE_ORTHOGRAPHY_JA", sr.GetID())
	// getDescription has no $match placeholder → SpecificIdRule keeps desc_spelling surface
	require.Equal(t, orthographyDefaultDescSpelling, sr.GetDescription())
}
