package disambiguation

// Twin of MultiWordChunkerTest (core)
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func hasPOS(sent *languagetool.AnalyzedSentence, want string) bool {
	for _, tok := range sent.GetTokens() {
		for _, r := range tok.GetReadings() {
			if r.GetPOSTag() != nil && *r.GetPOSTag() == want {
				return true
			}
		}
	}
	return false
}

func TestMultiWordChunker_languagetool_core_Disambiguate1(t *testing.T) {
	c := NewMultiWordChunker([]string{"New York\tB-NP"}, MultiWordChunkerSettings{AllowFirstCapitalized: true})
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.True(t, hasPOS(out, "B-NP") || hasPOS(out, "<B-NP>") || hasPOS(out, "</B-NP>"))
	// non-match sentence
	toks2 := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Los", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Angeles", nil, nil)),
	}
	out2 := c.Disambiguate(languagetool.NewAnalyzedSentence(toks2))
	require.False(t, hasPOS(out2, "B-NP") || hasPOS(out2, "<B-NP>"))
}

func TestMultiWordChunker_languagetool_core_Disambiguate2(t *testing.T) {
	c := NewMultiWordChunker2([]string{"New York\tB-NP"}, true)
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.True(t, hasPOS(out, "<B-NP>") || hasPOS(out, "B-NP"))
}

func TestMultiWordChunker_languagetool_core_Disambiguate2NoMatch(t *testing.T) {
	c := NewMultiWordChunker2([]string{"New York\tB-NP"}, true)
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Old", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.False(t, hasPOS(out, "<B-NP>"))
}

func TestMultiWordChunker_languagetool_core_Disambiguate2RemoveOtherReadings(t *testing.T) {
	c := NewMultiWordChunker2([]string{"New York\tB-NP"}, true)
	c.SetRemoveOtherReadings(true)
	c.SetWrapTag(true)
	tag := languagetool.SentenceStartTagName
	nn := "NN"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", &nn, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", &nn, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.True(t, hasPOS(out, "<B-NP>") || hasPOS(out, "B-NP"))
}

// Twin of MultiWordChunkerTest.testLettercaseVariants
// getInstance(..., true, true, true) → allowFirstCapitalized, allowAllUppercase, allowTitlecase
func TestMultiWordChunker_languagetool_core_LettercaseVariants(t *testing.T) {
	c := NewMultiWordChunker(nil, MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
		AllowAllUppercase:     true,
		AllowTitlecase:        true,
	})
	posRnB := "NCMS000_"
	lemmaRnB := "rhythm and blues"
	posVenus := "NCFSS00_"
	lemmaVenus := "Vênus de Milo"
	m := map[string]*languagetool.AnalyzedToken{
		"rhythm and blues": languagetool.NewAnalyzedToken("rhythm and blues", &posRnB, &lemmaRnB),
		"Vênus de Milo":    languagetool.NewAnalyzedToken("Vênus de Milo", &posVenus, &lemmaVenus),
	}
	rnb := c.GetTokenLettercaseVariants("rhythm and blues", m)
	require.Contains(t, rnb, "Rhythm and blues") // first-word upcase
	require.Contains(t, rnb, "Rhythm And Blues") // naïve titlecase
	require.Contains(t, rnb, "Rhythm and Blues") // smart titlecase (and exception)
	require.Contains(t, rnb, "RHYTHM AND BLUES") // all caps

	venus := c.GetTokenLettercaseVariants("Vênus de Milo", m)
	require.NotContains(t, venus, "Vênus De Milo") // naïve titlecase blocked (not all-lowercase)
	require.NotContains(t, venus, "vênus de milo") // downcased never generated
	require.Contains(t, venus, "VÊNUS DE MILO")    // all caps
}
