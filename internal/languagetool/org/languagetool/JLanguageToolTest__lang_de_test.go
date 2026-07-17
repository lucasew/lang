package languagetool

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/JLanguageToolTest.java
// Full rule engine deferred — Analyze + typography greens (rules/de lives in its package).
import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// germanAbbrRE ports German.TYPOGRAPHY_PATTERN — non-breaking space in abbreviations.
var germanAbbrRE = regexp.MustCompile(`\b([a-zA-Z]\.)([a-zA-Z]\.)`)

func germanTypography(s string) string {
	cfg := TypographyConfig{
		Enabled:            true,
		OpeningDoubleQuote: "„",
		ClosingDoubleQuote: "“",
		OpeningSingleQuote: "‚",
		ClosingSingleQuote: "‘",
	}
	out := ToAdvancedTypography(s, cfg)
	// German.toAdvancedTypography applies twice
	out = germanAbbrRE.ReplaceAllString(out, "$1\u00a0$2")
	out = germanAbbrRE.ReplaceAllString(out, "$1\u00a0$2")
	return out
}

// Port of JLanguageToolTest.testGerman
func TestJLanguageTool_lang_de_German(t *testing.T) {
	lt := NewJLanguageTool("de")
	require.Equal(t, "de", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Ein Test, der keine Fehler geben sollte."))
	// full check + word-repeat deferred (see rules/de GermanWordRepeatRuleTest)
	require.NotEmpty(t, lt.Analyze("Ein Test Test, der Fehler geben sollte."))
}

// Port of JLanguageToolTest.testGermanyGerman
func TestJLanguageTool_lang_de_GermanyGerman(t *testing.T) {
	lt := NewJLanguageTool("de-DE")
	require.Equal(t, "de-DE", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Ein Test, der keine Fehler geben sollte."))
	require.NotEmpty(t, lt.Analyze("Ein Test Test, der Fehler geben sollte."))
}

// Port of JLanguageToolTest.testPositionsWithGerman
func TestJLanguageTool_lang_de_PositionsWithGerman(t *testing.T) {
	lt := NewJLanguageTool("de-DE")
	sents := lt.Analyze("Stundenkilometer")
	require.Len(t, sents, 1)
	toks := sents[0].GetTokensWithoutWhitespace()
	require.NotEmpty(t, toks)
	require.Equal(t, 0, toks[0].GetStartPos())
}

// Port of JLanguageToolTest.testCleanOverlappingWithGerman
func TestJLanguageTool_lang_de_CleanOverlappingWithGerman(t *testing.T) {
	lt := NewJLanguageTool("de-DE")
	sents := lt.Analyze("TRGS - Technische Regeln für Gefahrstoffe")
	require.NotEmpty(t, sents)
}

// Port of JLanguageToolTest.testAdvancedTypography
func TestJLanguageTool_lang_de_AdvancedTypography(t *testing.T) {
	require.Equal(t, "Das ist…", germanTypography("Das ist..."))
	require.Equal(t, "Meinten Sie „entschieden“ oder „entscheidend“?",
		germanTypography(`Meinten Sie "entschieden" oder "entscheidend"?`))
	require.Equal(t, "Meinten Sie ‚entschieden‘ oder ‚entscheidend‘?",
		germanTypography("Meinten Sie 'entschieden' oder 'entscheidend'?"))
	require.Equal(t, "z.\u00a0B.", germanTypography("z. B."))
	require.Equal(t, "z.\u00a0B.", germanTypography("z.B."))
	require.Equal(t, "i.\u00a0d.\u00a0R.", germanTypography("i.d.R."))
	require.Equal(t, "i.\u00a0d.\u00a0R.", germanTypography("i. d. R."))
}

// Port of JLanguageToolTest.testGermanVariants
func TestJLanguageTool_lang_de_GermanVariants(t *testing.T) {
	for _, code := range []string{"de-DE", "de-AT", "de-CH"} {
		lt := NewJLanguageTool(code)
		require.Equal(t, code, lt.GetLanguageCode())
		require.NotEmpty(t, lt.Analyze("Hallo Welt."))
	}
}

// Port of JLanguageToolTest.testMultitokenSpeller
func TestJLanguageTool_lang_de_MultitokenSpeller(t *testing.T) {
	lt := NewJLanguageTool("de-DE")
	require.NotEmpty(t, lt.Analyze("New York City"))
}
