package de

// Twin of GermanConfusionProbabilityRuleTest.
// Java only has testConstructor (FakeLanguageModel + de); extra cases assert
// Java-visible constructor meta (exceptions, anti-patterns, filenames, id).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

// TestGermanConfusionProbabilityRule_Constructor twins
// GermanConfusionProbabilityRuleTest.testConstructor:
// constructing with FakeLanguageModel + de does not throw; id CONFUSION_RULE
// (Java GermanConfusionProbabilityRule does not override getId).
func TestGermanConfusionProbabilityRule_Constructor(t *testing.T) {
	// Java: new GermanConfusionProbabilityRule(TestTools.getEnglishMessages(),
	//   new FakeLanguageModel(), Languages.getLanguageForShortCode("de"));
	lm := ngrams.UniformLanguageModel(0.5, 1) // FakeLanguageModel stand-in
	r := NewGermanConfusionProbabilityRule(lm)
	require.NotNil(t, r)
	require.Equal(t, ngrams.ConfusionRuleID, r.GetID())
	require.Equal(t, "CONFUSION_RULE", r.GetID())

	// Java-visible constructor meta (not invent)
	require.Equal(t, []string{"confusion_sets.txt"}, r.GetFilenames())
	require.Len(t, r.Exceptions, 95)
	require.Len(t, GermanConfusionAntiPatterns, 11)
	require.Len(t, deConfusionSentenceExceptions, 11)
	require.NotNil(t, r.IsException)
	require.NotNil(t, r.IsCoveredByAntiPattern)
	require.NotNil(t, r.IsCommonWord)

	// Example pair mir/mit
	inex := r.GetIncorrectExamples()
	require.NotEmpty(t, inex)
	require.Contains(t, inex[0].GetExample(), "mir")
	require.Equal(t, []string{"mit"}, inex[0].GetCorrections())
	cex := r.GetCorrectExamples()
	require.NotEmpty(t, cex)
	require.Contains(t, cex[0].GetExample(), "mit")

	// Official confusion sets loaded (getFilenames path): inspiration resource/de
	// and/or embedded official copy. Fail-closed empty only if both absent.
	require.NotEmpty(t, r.WordToPairs, "constructor must load confusion_sets.txt like Java CPR")
}

func TestGermanConfusionProbabilityRule_ExceptionsCount(t *testing.T) {
	// Java GermanConfusionProbabilityRule.EXCEPTIONS length (comment-stripped string literals)
	require.Len(t, GermanConfusionExceptions, 95)
	require.Equal(t, "weist bei interesse auf", GermanConfusionExceptions[0])
	require.Equal(t, "Vorgestern und Gestern", GermanConfusionExceptions[len(GermanConfusionExceptions)-1])
	// Spot-check comment-adjacent entries (comments themselves are not exceptions)
	require.Contains(t, GermanConfusionExceptions, ", dir bei")
	require.Contains(t, GermanConfusionExceptions, "setz dir")
	require.NotContains(t, GermanConfusionExceptions, "froh, dir bei deiner Arbeit zu helfen")
	require.NotContains(t, GermanConfusionExceptions, "Setz dir doch bitte einen Termin am Donnerstag")
	// Wired into constructor
	r := NewGermanConfusionProbabilityRule(nil)
	require.Len(t, r.Exceptions, 95)
	require.True(t, r.IsLocalException("seht zu, dass wir gehen"))
}

func TestGermanConfusionAntiPatternsCount(t *testing.T) {
	// Java GermanConfusionProbabilityRule.ANTI_PATTERNS length
	require.Len(t, GermanConfusionAntiPatterns, 11)
	require.NotNil(t, NewGermanConfusionProbabilityRule(nil).IsCoveredByAntiPattern)
}

func TestGermanConfusionProbabilityRule_SentenceException(t *testing.T) {
	lm := ngrams.UniformLanguageModel(0.5, 1)
	r := NewGermanConfusionProbabilityRule(lm)
	// Sentence exception pattern "wir (" should skip via IsException.
	require.True(t, r.IsException("Hallo, wir (die Dingsbums Gmbh)", 0, 3))
	// deConfusionIsException also
	require.True(t, deConfusionIsException("Wie heizen Sie das Haus?", 0, 3))
	// Java: fiel(e|en)? .* (aus|auf|anheim) — needs intermediate token(s)
	require.True(t, deConfusionIsException("fiel etwas aus dem Rahmen", 0, 4))
	require.False(t, deConfusionIsException("Ganz normaler Satz ohne Ausnahme.", 0, 4))
}

func TestGermanConfusionProbabilityRule_IsCommonWord(t *testing.T) {
	// Java COMMON_WORD = [\\wöäüßÖÄÜ]+ with Matcher.matches()
	r := NewGermanConfusionProbabilityRule(nil)
	require.True(t, r.IsCommonWord("Wörter"))
	require.True(t, r.IsCommonWord("größer"))
	require.True(t, r.IsCommonWord("abc"))
	require.False(t, r.IsCommonWord("foo-bar"))
	require.False(t, r.IsCommonWord("a b"))
	require.False(t, r.IsCommonWord(""))
}

func TestGermanConfusionAntiPatternImmunizeFasstZusammen(t *testing.T) {
	// token-only anti-pattern: fasst … zusammen
	toks := withPositions(
		sentStartATR(),
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "der"),
		atrWithPOS("Übersicht", "SUB:NOM:SIN:FEM", "Übersicht"),
		atrWithPOS("fasst", "VER:3:SIN:PRS:NON", "fassen"),
		atrWithPOS("Ziele", "SUB:AKK:PLU:NEU", "Ziel"),
		atrWithPOS("zusammen", "ZUS", "zusammen"),
		atrWithPOS(".", "PKT", "."),
	)
	sent := languagetool.NewAnalyzedSentence(toks)
	fasst := toks[3]
	require.True(t, deConfusionIsCoveredByAntiPattern(sent, fasst.GetStartPos(), fasst.GetEndPos()))
}

func TestGermanConfusionProbabilityRule_WithLMAndPair(t *testing.T) {
	// Fake LM: always prefer "mit" over "mir" in 3-grams
	lm := ngrams.FuncLanguageModel(func(tokens []string) ngrams.Probability {
		for _, tok := range tokens {
			if tok == "mit" {
				return ngrams.NewProbabilitySimple(0.9, 1.0)
			}
			if tok == "mir" {
				return ngrams.NewProbabilitySimple(0.001, 1.0)
			}
		}
		return ngrams.NewProbabilitySimple(0.1, 1.0)
	})
	r := NewGermanConfusionProbabilityRule(lm)
	// Force pair mir/mit for the test (like Java setConfusionPair).
	pair := rules.NewConfusionPairTokens("mir", "mit", 10, true)
	r.SetConfusionPair(pair)
	matches := r.Match(languagetool.AnalyzePlain("Während Sie das Ganze mir einem Holzlöffel rühren."))
	require.NotEmpty(t, matches)
	require.Contains(t, matches[0].GetSuggestedReplacements(), "mit")
	// SpecificIdRule id prefix uses getId() → CONFUSION_RULE_… (not DE_CONFUSION_RULE)
	id := matches[0].GetSpecificRuleId()
	if id == "" {
		// fallback: base rule id when specific id unset
		id = r.GetID()
	}
	require.Contains(t, id, "CONFUSION_RULE")
	require.NotContains(t, id, "DE_CONFUSION_RULE")
}

func TestGermanConfusion_ExceptionPhraseSkips(t *testing.T) {
	// "so viel" is in EXCEPTIONS — should not fire when phrase covers span
	lm := ngrams.FuncLanguageModel(func(tokens []string) ngrams.Probability {
		for _, tok := range tokens {
			if tok == "fiel" {
				return ngrams.NewProbabilitySimple(0.9, 1.0)
			}
			if tok == "viel" {
				return ngrams.NewProbabilitySimple(0.001, 1.0)
			}
		}
		return ngrams.NewProbabilitySimple(0.1, 1.0)
	})
	r := NewGermanConfusionProbabilityRule(lm)
	pair := rules.NewConfusionPairTokens("viel", "fiel", 10, true)
	r.SetConfusionPair(pair)
	matches := r.Match(languagetool.AnalyzePlain("Es gibt so viel zu tun."))
	require.Empty(t, matches)
}

func TestGermanConfusion_LoadsOfficialConfusionSets(t *testing.T) {
	// Java CPR loads /de/confusion_sets.txt in constructor via getFilenames().
	// Prefer inspiration resource path when present (same as DiscoverGermanResourceDir).
	if p := discoverDEConfusionSet("confusion_sets.txt"); p != "" {
		require.FileExists(t, p)
	}
	r := NewGermanConfusionProbabilityRule(ngrams.UniformLanguageModel(0.5, 1))
	require.NotEmpty(t, r.WordToPairs,
		"WordToPairs must load from official de/confusion_sets.txt (fail-closed empty only if resource missing)")
	// Spot-check keys from official DE confusion_sets (mir/mit is the example pair)
	hasMir := len(r.WordToPairs["mir"]) > 0 || len(r.WordToPairs["mit"]) > 0
	hasSeit := len(r.WordToPairs["seit"]) > 0 || len(r.WordToPairs["seid"]) > 0
	require.True(t, hasMir || hasSeit, "expected at least one official confusion pair key")
}

func TestGermanConfusion_WithoutLMNoInvent(t *testing.T) {
	r := NewGermanConfusionProbabilityRule(nil)
	require.Equal(t, "CONFUSION_RULE", r.GetID())
	// Without LM, Match must not invent hits even with official pairs loaded
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Während Sie das Ganze mir einem Holzlöffel rühren.")))
}
