package de

// Twin of GermanDisambiguationTest — chunker + MultitokenIgnore from official lists
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
	disde "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/de"
	"github.com/stretchr/testify/require"
)

func atrDE(token, pos string) *languagetool.AnalyzedTokenReadings {
	p := pos
	return languagetool.NewAnalyzedTokenReadingsList(
		[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(token, &p, nil)}, 0)
}

func sentStartDE() *languagetool.AnalyzedTokenReadings {
	// SENT_START so MultiWordChunker space branch (j > 1) matches Java indexing.
	tag := languagetool.SentenceStartTagName
	return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil))
}

// multiwordContent builds SENT_START + token + " " + token for MultiWordChunker.
func multiwordContent(a, b string) []*languagetool.AnalyzedTokenReadings {
	return []*languagetool.AnalyzedTokenReadings{
		sentStartDE(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(a, nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(b, nil, nil)),
	}
}

// requireIgnoredBoth asserts isIgnoredBySpeller on both content tokens (Java tokens[1], tokens[2]).
func requireIgnoredBoth(t *testing.T, d *disde.GermanRuleDisambiguator, a, b string, want bool) {
	t.Helper()
	out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordContent(a, b)))
	toks := out.GetTokens()
	require.Len(t, toks, 4)
	if want {
		require.True(t, toks[1].IsIgnoredBySpeller(), "%q %q first content token", a, b)
		require.True(t, toks[3].IsIgnoredBySpeller(), "%q %q second content token", a, b)
	} else {
		require.False(t, toks[1].IsIgnoredBySpeller(), "%q %q first content token should NOT ignore", a, b)
		require.False(t, toks[3].IsIgnoredBySpeller(), "%q %q second content token should NOT ignore", a, b)
	}
}

// Port of GermanDisambiguationTest.testChunker
func TestGermanDisambiguation_Chunker(t *testing.T) {
	// GermanChunker NP tags (Java same test method has chunk cases).
	// Full analyzeText POS strings need GermanTagger/dict pipeline — not claimed here.
	toks2 := []*languagetool.AnalyzedTokenReadings{
		atrDE("ein", "ART:IND:AKK:SIN:NEU"),
		atrDE("Konzept", "SUB:AKK:SIN:NEU"),
	}
	ch := chunking.NewGermanChunker()
	ch.AddChunkTags(toks2)
	// REGEXES2 may add NPS on top of B-NP/I-NP (Java additive tags).
	require.Contains(t, toks2[0].GetChunkTags(), "B-NP")
	require.Contains(t, toks2[1].GetChunkTags(), "I-NP")

	// Java MultitokenIgnore via GermanRuleDisambiguator.multitokenSpeller
	// (/de/multitoken-ignore.txt, tagForNotAddingTags, setIgnoreSpelling true).
	if disde.DiscoverGermanMultitokenIgnore() == "" {
		t.Skip("official multitoken-ignore.txt not discoverable")
	}
	d := disde.NewGermanRuleDisambiguator()
	require.NotNil(t, d.MultitokenIgnore, "NewGermanRuleDisambiguator wires MultitokenIgnore")

	// Official entries:
	//   3-adische System/E
	//   3-adischen System/S
	//   3-adischen Systeme/N
	//   Kelassurier Mauer/N
	// Java GermanDisambiguationTest.testChunker isIgnoredBySpeller cases:

	// "3-adische System" → both content tokens ignored
	requireIgnoredBoth(t, d, "3-adische", "System", true)
	// "3-adische Systeme" → ignored (/E expansion)
	requireIgnoredBoth(t, d, "3-adische", "Systeme", true)
	// "3-adischen Systems" → ignored (/S)
	requireIgnoredBoth(t, d, "3-adischen", "Systems", true)

	// "Kelassurier Mauer" → ignored
	requireIgnoredBoth(t, d, "Kelassurier", "Mauer", true)
	// "Kelassurier Mauern" → ignored (/N)
	requireIgnoredBoth(t, d, "Kelassurier", "Mauern", true)
	// "Kelassurier Mauers" → not ignored (no /S on Mauer)
	requireIgnoredBoth(t, d, "Kelassurier", "Mauers", false)

	// Bare Kelassurier alone is not a multitoken match — no invent single-token ignore.
	toksBare := []*languagetool.AnalyzedTokenReadings{
		sentStartDE(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Kelassurier", nil, nil)),
	}
	outBare := d.Disambiguate(languagetool.NewAnalyzedSentence(toksBare))
	require.False(t, outBare.GetTokens()[1].IsIgnoredBySpeller(), "single Kelassurier needs multiword partner")

	// Optional confidence: official "Smart Home/S" (not invent).
	requireIgnoredBoth(t, d, "Smart", "Home", true)
	requireIgnoredBoth(t, d, "Smart", "Homes", true)
}
