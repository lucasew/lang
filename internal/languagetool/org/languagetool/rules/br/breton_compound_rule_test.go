package br

// Outcome twin for org.languagetool.rules.br.BretonCompoundRule
// (no dedicated Java *CompoundRuleTest for BR — RO-parity outcome matrix).
//
// Java ctor metadata:
//   id BR_COMPOUNDS, description "Mots composés"
//   Categories.COMPOUNDING, ITSIssueType.Grammar
//   4 Breton messages + shortMessage
//   example: alc'hweder gwez → alc'hweder-gwez
//   getCompoundRuleData → CompoundRuleData("/br/compounds.txt")
//
// Analysis path (same as MorfologikBretonSpellerRuleTest / JLanguageTool(new Breton())):
//   AnalyzeWithTokenizer + BretonWordTokenizer.
// Note: BretonWordTokenizer normalizes apostrophes to U+2019; compounds.txt keeps
// ASCII U+0027 in c'h entries. That mismatch is Java-identical — non-apostrophe
// entries exercise the real BR path; the official example entry is checked via
// data + a surface token join that preserves ASCII (AnalyzePlain), proving
// rule+data for the documented pair.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	brtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/br"
	"github.com/stretchr/testify/require"
)

// analyzeBR ports JLanguageTool(new Breton()).getAnalyzedSentence: BretonWordTokenizer.
func analyzeBRCompound(text string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTokenizer(text, brtok.NewBretonWordTokenizer())
}

func TestBretonCompoundRule_Metadata(t *testing.T) {
	r := NewBretonCompoundRule(nil)
	// Java getId / getDescription
	require.Equal(t, "BR_COMPOUNDS", r.GetID())
	require.Equal(t, "Mots composés", r.GetDescription())
	// Java: super.setCategory(Categories.COMPOUNDING.getCategory(messages))
	require.NotNil(t, r.GetCategory())
	require.Equal(t, "COMPOUNDING", r.GetCategory().GetID().String())
	// Java: setLocQualityIssueType(ITSIssueType.Grammar)
	require.Equal(t, rules.ITSGrammar, r.GetLocQualityIssueType())
	// Java 4-arg super shortMessage + BR messages
	require.Equal(t, "Kudenn barrennig-stagañ", r.ShortDesc)
	require.Equal(t, "Skrivet e vez ar ger-mañ boaz gant ur varrennig-stagañ.", r.WithHyphenMessage)
	require.Equal(t, "Ar ger-mañ a zo skrivet boaz evel unan hepken.", r.WithoutHyphenMessage)
	require.Equal(t, "An droienn-mañ a zo skrivet evel ur ger hepken pe gant ur varrennig-stagañ.", r.WithOrWithoutHyphenMessage)
	// Java AbstractCompoundRule default / field true
	require.True(t, r.SentenceStartsWithUpperCase)
	// Java: no useSubRuleSpecificIds()
	require.False(t, r.SubRuleSpecificIDs)
	// Java: no getUrl override
	require.Equal(t, "", r.GetURL())
	// Java getCompoundRuleData singleton from /br/compounds.txt
	require.Equal(t, BretonCompoundFilePath, "/br/compounds.txt")
	require.NotNil(t, r.GetCompoundRuleData())
	require.Same(t, loadCompoundData(), r.GetCompoundRuleData())
	require.Same(t, r.Data, r.GetCompoundRuleData())
	// data contains official example entry (ASCII apostrophe as in compounds.txt)
	require.True(t, r.GetCompoundRuleData().ContainsIncorrect("alc'hweder gwez"))
	require.True(t, r.GetCompoundRuleData().ContainsDash("alc'hweder gwez"))
	// Java example pair
	inex := r.GetIncorrectExamples()
	require.Len(t, inex, 1)
	require.Equal(t, []string{"alc'hweder-gwez"}, inex[0].GetCorrections())
	require.Contains(t, inex[0].GetExample(), "alc'hweder gwez")
	cex := r.GetCorrectExamples()
	require.Len(t, cex, 1)
	require.Contains(t, cex[0].GetExample(), "alc'hweder-gwez")
}

func TestBretonCompoundRule_Rule(t *testing.T) {
	// All official BR compounds.txt entries end with '*' → dash-only → WithHyphenMessage.
	rule := NewBretonCompoundRule(nil)
	withHyphenMsg := "Skrivet e vez ar ger-mañ boaz gant ur varrennig-stagañ."
	shortMsg := "Kudenn barrennig-stagañ"

	check := func(analyze func(string) *languagetool.AnalyzedSentence, expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		sent := analyze(text)
		require.NotNil(t, sent, "analyze(%q)", text)
		matches := rule.Match(sent)
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if len(expSuggestions) > 0 {
			require.Equal(t, 1, expectedErrors)
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements(), "text %q", text)
			require.Equal(t, withHyphenMsg, matches[0].GetMessage(), "text %q", text)
			require.Equal(t, shortMsg, matches[0].GetShortMessage(), "text %q", text)
			// no useSubRuleSpecificIds → specific id falls back to rule id
			require.Equal(t, "BR_COMPOUNDS", matches[0].GetSpecificRuleId())
		}
	}

	// --- Real BR analysis path (BretonWordTokenizer), non-apostrophe compounds ---
	// abardaez-noz*
	check(analyzeBRCompound, 1, "abardaez noz", "abardaez-noz")
	check(analyzeBRCompound, 0, "abardaez-noz")
	// sentence-start capitalisation (SentenceStartsWithUpperCase=true)
	check(analyzeBRCompound, 1, "Abardaez noz", "Abardaez-noz")
	// a-bouez-penn* (3 parts)
	check(analyzeBRCompound, 1, "a bouez penn", "a-bouez-penn")
	check(analyzeBRCompound, 0, "a-bouez-penn")
	// dibenn-sizhun*
	check(analyzeBRCompound, 1, "dibenn sizhun", "dibenn-sizhun")
	check(analyzeBRCompound, 0, "dibenn-sizhun")

	// --- Official example pair (ASCII apostrophe in compounds.txt) ---
	// AnalyzePlain joins letter/punct tokens without inserting spaces, so the
	// surface key "alc'hweder gwez" matches data (proves rule+data for the pair).
	// BretonWordTokenizer maps ' → U+2019 and would not hit the ASCII data key
	// (Java-identical tokenizer behaviour — not a Go invent).
	check(languagetool.AnalyzePlain, 1, "alc'hweder gwez", "alc'hweder-gwez")
	check(languagetool.AnalyzePlain, 0, "alc'hweder-gwez")
	check(languagetool.AnalyzePlain, 1, "Alc'hweder gwez", "Alc'hweder-gwez")
	check(languagetool.AnalyzePlain, 1, "un alc'hweder gwez e-kerzh an dibenn-sizhun", "alc'hweder-gwez")
	check(languagetool.AnalyzePlain, 0, "un alc'hweder-gwez e-kerzh an dibenn-sizhun")
}
