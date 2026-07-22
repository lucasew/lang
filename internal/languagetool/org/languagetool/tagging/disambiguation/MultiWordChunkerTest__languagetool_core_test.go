package disambiguation

// Twin of MultiWordChunkerTest (languagetool-core)
// Java: inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/tagging/disambiguation/MultiWordChunkerTest.java
// Resource: .../src/test/resources/org/languagetool/resource/yy/multiwords.txt
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// hasPOS is shared with the standalone twin tests (loose POS presence helper).
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

// yyMultiwordsPath resolves Java test resource /yy/multiwords.txt under inspiration.
func yyMultiwordsPath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
	p := filepath.Join(dir,
		"inspiration/languagetool/languagetool-core/src/test/resources/org/languagetool/resource/yy/multiwords.txt")
	_, err = os.Stat(p)
	require.NoError(t, err, "Java yy multiwords resource must exist")
	return p
}

func loadYYMultiWordChunker(t *testing.T) *MultiWordChunker {
	t.Helper()
	f, err := os.Open(yyMultiwordsPath(t))
	require.NoError(t, err)
	defer f.Close()
	// Java: MultiWordChunker.getInstance("/yy/multiwords.txt", true, true, true)
	c, err := NewMultiWordChunkerFromReader(f, MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
		AllowAllUppercase:     true,
		AllowTitlecase:        true,
	})
	require.NoError(t, err)
	return c
}

func loadYYMultiWordChunker2(t *testing.T) *MultiWordChunker2 {
	t.Helper()
	f, err := os.Open(yyMultiwordsPath(t))
	require.NoError(t, err)
	defer f.Close()
	lines, err := loadMultiWordLines(f)
	require.NoError(t, err)
	// Java: new MultiWordChunker2("/yy/multiwords.txt", true)
	return NewMultiWordChunker2(lines, true)
}

// fakeAnalyzedSentence ports Java MultiWordChunkerTest setup:
// FakeLanguage + DemoTagger that injects FakePosTag on non-whitespace tokens,
// then JLanguageTool.getAnalyzedSentence(text).
func fakeAnalyzedSentence(text string) *languagetool.AnalyzedSentence {
	// FakeLanguage uses default WordTokenizer (AnalyzePlain).
	sent := languagetool.AnalyzePlain(text)
	fakePOS := "FakePosTag"
	for _, readings := range sent.GetTokens() {
		if readings == nil || readings.IsWhitespace() || readings.IsSentenceStart() {
			continue
		}
		tok := readings.GetToken()
		// Java: new AnalyzedToken(readings.getToken(), "FakePosTag", readings.getToken())
		readings.AddReading(languagetool.NewAnalyzedToken(tok, &fakePOS, &tok), "")
	}
	return sent
}

// readingsListString ports Java tokens[i].getReadings().toString() (List.toString of AnalyzedToken).
func readingsListString(tok *languagetool.AnalyzedTokenReadings) string {
	if tok == nil {
		return "[]"
	}
	parts := make([]string, 0, len(tok.GetReadings()))
	for _, r := range tok.GetReadings() {
		if r == nil {
			parts = append(parts, "null")
			continue
		}
		parts = append(parts, r.String())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// Port of MultiWordChunkerTest.testDisambiguate1
func TestMultiWordChunker_languagetool_core_Disambiguate1(t *testing.T) {
	multiWordChunker := loadYYMultiWordChunker(t)

	analyzedSentence := fakeAnalyzedSentence("ah for shame")
	disambiguated := multiWordChunker.Disambiguate(analyzedSentence)
	tokens := disambiguated.GetTokens()
	require.GreaterOrEqual(t, len(tokens), 6)

	// tokens: [0]=SENT_START [1]=ah [2]=sp [3]=for [4]=sp [5]=shame
	require.True(t, strings.Contains(readingsListString(tokens[1]), "<adv>"),
		"tokens[1] readings=%s", readingsListString(tokens[1]))
	require.False(t, strings.Contains(readingsListString(tokens[3]), "adv"),
		"tokens[3] readings=%s", readingsListString(tokens[3]))
	require.True(t, strings.Contains(readingsListString(tokens[5]), "</adv>"),
		"tokens[5] readings=%s", readingsListString(tokens[5]))

	require.True(t, strings.Contains(readingsListString(tokens[1]), "FakePosTag"))
	require.True(t, strings.Contains(readingsListString(tokens[3]), "FakePosTag"))
	require.True(t, strings.Contains(readingsListString(tokens[5]), "FakePosTag"))
}

// Port of MultiWordChunkerTest.testDisambiguate2
func TestMultiWordChunker_languagetool_core_Disambiguate2(t *testing.T) {
	multiWordChunker := loadYYMultiWordChunker2(t)

	analyzedSentence := fakeAnalyzedSentence("Ah for shame")
	disambiguated := multiWordChunker.Disambiguate(analyzedSentence)
	tokens := disambiguated.GetTokens()
	require.GreaterOrEqual(t, len(tokens), 6)

	require.True(t, strings.Contains(readingsListString(tokens[1]), "<adv>"),
		"tokens[1] readings=%s", readingsListString(tokens[1]))
	require.True(t, strings.Contains(readingsListString(tokens[3]), "<adv>"),
		"tokens[3] readings=%s", readingsListString(tokens[3]))
	require.True(t, strings.Contains(readingsListString(tokens[5]), "<adv>"),
		"tokens[5] readings=%s", readingsListString(tokens[5]))

	require.True(t, strings.Contains(readingsListString(tokens[1]), "FakePosTag"))
	require.True(t, strings.Contains(readingsListString(tokens[3]), "FakePosTag"))
	require.True(t, strings.Contains(readingsListString(tokens[5]), "FakePosTag"))
}

// Port of MultiWordChunkerTest.testDisambiguate2NoMatch
func TestMultiWordChunker_languagetool_core_Disambiguate2NoMatch(t *testing.T) {
	multiWordChunker := loadYYMultiWordChunker2(t)

	analyzedSentence := fakeAnalyzedSentence("ahh for shame")
	disambiguated := multiWordChunker.Disambiguate(analyzedSentence)
	tokens := disambiguated.GetTokens()
	require.GreaterOrEqual(t, len(tokens), 2)

	require.False(t, strings.Contains(readingsListString(tokens[1]), "<adv>"),
		"tokens[1] readings=%s", readingsListString(tokens[1]))
}

// Port of MultiWordChunkerTest.testDisambiguate2RemoveOtherReadings
func TestMultiWordChunker_languagetool_core_Disambiguate2RemoveOtherReadings(t *testing.T) {
	multiWordChunker := loadYYMultiWordChunker2(t)
	// Java: setRemoveOtherReadings(true); setWrapTag(false);
	multiWordChunker.SetRemoveOtherReadings(true)
	multiWordChunker.SetWrapTag(false)

	analyzedSentence := fakeAnalyzedSentence("ah for shame")
	disambiguated := multiWordChunker.Disambiguate(analyzedSentence)
	tokens := disambiguated.GetTokens()
	require.GreaterOrEqual(t, len(tokens), 6)

	require.True(t, strings.Contains(readingsListString(tokens[1]), "adv"),
		"tokens[1] readings=%s", readingsListString(tokens[1]))
	require.True(t, strings.Contains(readingsListString(tokens[3]), "adv"),
		"tokens[3] readings=%s", readingsListString(tokens[3]))
	require.True(t, strings.Contains(readingsListString(tokens[5]), "adv"),
		"tokens[5] readings=%s", readingsListString(tokens[5]))

	require.False(t, strings.Contains(readingsListString(tokens[1]), "FakePosTag"),
		"tokens[1] readings=%s", readingsListString(tokens[1]))
	require.False(t, strings.Contains(readingsListString(tokens[3]), "FakePosTag"),
		"tokens[3] readings=%s", readingsListString(tokens[3]))
	require.False(t, strings.Contains(readingsListString(tokens[5]), "FakePosTag"),
		"tokens[5] readings=%s", readingsListString(tokens[5]))
}

// Port of MultiWordChunkerTest.testLettercaseVariants
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
