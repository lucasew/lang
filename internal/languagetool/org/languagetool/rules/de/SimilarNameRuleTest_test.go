package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/SimilarNameRuleTest.java
//
// Java:
//   SimilarNameRule rule = new SimilarNameRule(TestTools.getEnglishMessages());
//   JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("de-DE"));
//   AnalyzedSentence sentence = lt.getAnalyzedSentence(input);
//   RuleMatch[] matches = rule.match(Collections.singletonList(sentence));
//
// Go: de-DE JLanguageTool + WireGermanTagWord (createDefaultTagger) so EIG:/isPosTagUnknown
// match real GermanTagger — no surface invent for name detection.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// newSimilarNameFixture mirrors Java testRule setup (rule + de-DE analysis).
func newSimilarNameFixture(t *testing.T) (*SimilarNameRule, *languagetool.JLanguageTool) {
	t.Helper()
	rule := NewSimilarNameRule(nil) // Java: TestTools.getEnglishMessages() — category msgs unused for DE_SIMILAR_NAMES
	lt := languagetool.NewJLanguageTool("de-DE")
	// Java createDefaultTagger → GermanTagger; without TagWord all tokens are isPosTagUnknown.
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for SimilarNameRule twin")
	return rule, lt
}

// TestSimilarNameRule_Rule ports SimilarNameRuleTest.testRule full matrix.
func TestSimilarNameRule_Rule(t *testing.T) {
	rule, lt := newSimilarNameFixture(t)

	assertErrors := func(input string, expectedMatches int) {
		t.Helper()
		// Java: rule.match(Collections.singletonList(lt.getAnalyzedSentence(input)))
		sentence := lt.GetAnalyzedSentence(input)
		ms := rule.MatchList([]*languagetool.AnalyzedSentence{sentence})
		require.Equal(t, expectedMatches, len(ms), "input=%q", input)
		if expectedMatches == 1 {
			require.Equal(t, "DE_SIMILAR_NAMES", ms[0].GetRule().(interface{ GetID() string }).GetID())
		}
	}

	// Java SimilarNameRuleTest.testRule matrix:
	assertErrors("Hier steht Angela Müller. Im nächsten Satz dann Miller.", 1)
	assertErrors("Hier steht Angela Müller. Im nächsten Satz dann Müllers Ehemann.", 0)
	assertErrors("Hier steht Angela Müller. Dann Mulla, nicht ähnlich genug.", 0)
	assertErrors("Ein Mikrocontroller, bei Mikrocontrollern", 0)
	assertErrors("Hier steht das Rad Deiner Freundin. Und Deinem Hund geht es gut?", 0)
}

// Java-visible outcomes beyond match count: suggestion, span, message, meta.
func TestSimilarNameRule_SuggestionAndMeta(t *testing.T) {
	rule, lt := newSimilarNameFixture(t)

	sentence := lt.GetAnalyzedSentence("Hier steht Angela Müller. Im nächsten Satz dann Miller.")
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{sentence})
	require.Equal(t, 1, len(ms))
	require.Equal(t, "DE_SIMILAR_NAMES", ms[0].GetRule().(interface{ GetID() string }).GetID())
	// Java: ruleMatch.setSuggestedReplacement(similarName) → prior "Müller"
	require.Equal(t, []string{"Müller"}, ms[0].GetSuggestedReplacements())
	require.Contains(t, ms[0].GetMessage(), "Miller")
	require.Contains(t, ms[0].GetMessage(), "Müller")
	require.Contains(t, ms[0].GetMessage(), "Tippfehler")
	// span on Miller token (Java pos+token.getStartPos/EndPos; UTF-16 units, not Go UTF-8 bytes)
	var millerTok *languagetool.AnalyzedTokenReadings
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "Miller" {
			millerTok = tok
			break
		}
	}
	require.NotNil(t, millerTok)
	require.Equal(t, millerTok.GetStartPos(), ms[0].GetFromPos())
	require.Equal(t, millerTok.GetEndPos(), ms[0].GetToPos())
	require.Equal(t, 6, ms[0].GetToPos()-ms[0].GetFromPos()) // Miller UTF-16 length

	// constructor / TextLevelRule surface
	require.Equal(t, "DE_SIMILAR_NAMES", rule.GetID())
	require.Equal(t, "Mögliche Tippfehler in Namen finden", rule.GetDescription())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	require.Equal(t, -1, rule.EstimateContextForSureMatch())
	require.True(t, rule.IsDefaultOff())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Len(t, rule.GetIncorrectExamples(), 1)
	require.Len(t, rule.GetCorrectExamples(), 1)
	require.Equal(t, []string{"Müller"}, rule.GetIncorrectExamples()[0].GetCorrections())
	// compile-time: implements TextLevelRule
	var _ rules.TextLevelRule = rule
}

// isMaybeName unit gates (Java expression + startsWithUppercase).
// EIG inject documents POS path; real tagger covers integration above.
func TestSimilarNameRule_IsMaybeName(t *testing.T) {
	// Tagged non-EIG SUB must not be a name (Java: not EIG and not isPosTagUnknown)
	sub := atrWithPOS("Miller", "SUB:NOM:SIN:MAS", "Miller")
	require.False(t, isMaybeName(sub))

	// EIG without :COU → name
	eig := atrWithPOS("Miller", "EIG:NOM:SIN:MAS:NAC", "Miller")
	require.True(t, isMaybeName(eig))

	// EIG country :COU excluded
	cou := atrWithPOS("Spanien", "EIG:NOM:SIN:NEU:COU", "Spanien")
	require.False(t, isMaybeName(cou))

	// Single null-POS reading → isPosTagUnknown (AnalyzePlain path)
	unk := languagetool.AnalyzePlain("Miller").GetTokensWithoutWhitespace()
	found := false
	for _, tok := range unk {
		if tok != nil && tok.GetToken() == "Miller" {
			require.True(t, tok.IsPosTagUnknown())
			require.True(t, isMaybeName(tok))
			found = true
		}
	}
	require.True(t, found)

	// Exclude list (exact equals, even if unknown/uppercase)
	for _, w := range []string{"Dein", "Deine", "Deinen", "Deiner", "Deines", "Deinem",
		"Ihr", "Ihre", "Ihren", "Ihrer", "Ihres", "Ihrem"} {
		s := languagetool.AnalyzePlain(w)
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if tok != nil && tok.GetToken() == w {
				require.False(t, isMaybeName(tok), "exclude %q", w)
			}
		}
	}

	// length < MIN_LENGTH=4
	short := atrWithPOS("Max", "EIG:NOM:SIN:MAS", "Max")
	require.False(t, isMaybeName(short))

	// lowercase not name (startsWithUppercase gate)
	low := atrWithPOS("miller", "EIG:NOM:SIN:MAS", "miller")
	require.False(t, isMaybeName(low))
}

// similarName unit: equal skip, genitive s/n, lenDiff + Levenshtein ≤ MAX_DIFF.
func TestSimilarNameRule_SimilarName(t *testing.T) {
	r := NewSimilarNameRule(nil)
	names := map[string]struct{}{"Müller": {}}

	// equal → skip (no self-match)
	require.Equal(t, "", r.similarName("Müller", names))

	// genitive s: Müllers vs Müller → skip
	require.Equal(t, "", r.similarName("Müllers", names))
	// reverse asymmetry also skip
	require.Equal(t, "", r.similarName("Müller", map[string]struct{}{"Müllers": {}}))

	// dative-ish n asymmetry
	require.Equal(t, "", r.similarName("Merkeln", map[string]struct{}{"Merkel": {}}))

	// Miller ~ Müller (diff 1 on ü/i in UTF-16)
	require.Equal(t, "Müller", r.similarName("Miller", names))

	// Mulla: lenDiff=2 > MAX_DIFF → no
	require.Equal(t, "", r.similarName("Mulla", names))

	// Levenshtein BMP
	require.Equal(t, 1, levenshteinSimilar("Merkel", "Merkl"))
	require.Equal(t, 0, levenshteinSimilar("Haus", "Haus"))
}

// pos accumulates correctedTextLength across multi-sentence list (Java match loop).
func TestSimilarNameRule_MultiSentencePos(t *testing.T) {
	rule, lt := newSimilarNameFixture(t)
	s1 := lt.GetAnalyzedSentence("Hier steht Angela Müller.")
	s2 := lt.GetAnalyzedSentence("Im nächsten Satz dann Miller.")
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"Müller"}, ms[0].GetSuggestedReplacements())

	var millerStart int
	found := false
	for _, tok := range s2.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "Miller" {
			millerStart = tok.GetStartPos()
			found = true
			break
		}
	}
	require.True(t, found, "Miller token in s2")
	wantFrom := s1.GetCorrectedTextLength() + millerStart
	require.Equal(t, wantFrom, ms[0].GetFromPos())
	require.Equal(t, wantFrom+6, ms[0].GetToPos())
}
