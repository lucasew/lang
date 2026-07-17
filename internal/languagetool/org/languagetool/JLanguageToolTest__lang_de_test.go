package languagetool

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/JLanguageToolTest.java
// Check path with inject SimpleWordRepeatChecker; full rule stack still deferred.
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
	out = germanAbbrRE.ReplaceAllString(out, "$1\u00a0$2")
	out = germanAbbrRE.ReplaceAllString(out, "$1\u00a0$2")
	return out
}

func deLTWithRepeat() *JLanguageTool {
	lt := NewJLanguageTool("de")
	lt.AddChecker(SimpleWordRepeatChecker("GERMAN_WORD_REPEAT_RULE"))
	return lt
}

// Port of JLanguageToolTest.testGerman
func TestJLanguageTool_lang_de_German(t *testing.T) {
	lt := deLTWithRepeat()
	require.Equal(t, "de", lt.GetLanguageCode())
	require.Empty(t, lt.Check("Ein Test, der keine Fehler geben sollte."))
	require.Len(t, lt.Check("Ein Test Test, der Fehler geben sollte."), 1)

	// unknown words listing
	lt.SetListUnknownWords(true)
	lt.IsKnownWord = KnownWordSet("Ein", "Test", "der", "keine", "Fehler", "geben", "sollte")
	_ = lt.Check("I can give you more a detailed description")
	unk := lt.GetUnknownWords()
	require.NotEmpty(t, unk)
	require.Contains(t, unk, "I")
	require.Contains(t, unk, "can")
	require.Contains(t, unk, "description")
}

// Port of JLanguageToolTest.testGermanyGerman
func TestJLanguageTool_lang_de_GermanyGerman(t *testing.T) {
	lt := NewJLanguageTool("de-DE")
	lt.AddChecker(SimpleWordRepeatChecker("GERMAN_WORD_REPEAT_RULE"))
	require.Empty(t, lt.Check("Ein Test, der keine Fehler geben sollte."))
	require.NotEmpty(t, lt.Check("Ein Test Test, der Fehler geben sollte."))

	lt.SetListUnknownWords(true)
	lt.IsKnownWord = KnownWordSet("Ein", "Test", "der", "keine", "Fehler", "geben", "sollte")
	_ = lt.Check("I can give you more a detailed description")
	require.NotEmpty(t, lt.GetUnknownWords())
}

// Port of JLanguageToolTest.testPositionsWithGerman
func TestJLanguageTool_lang_de_PositionsWithGerman(t *testing.T) {
	lt := NewJLanguageTool("de-DE")
	// inject misspelling-like match on whole token via custom checker
	lt.AddChecker(func(s *AnalyzedSentence) []LocalMatch {
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if tok != nil && tok.GetToken() == "Stundenkilometer" {
				return []LocalMatch{{
					FromPos: tok.GetStartPos(),
					ToPos:   tok.GetEndPos(),
					RuleID:  "COMPOUND",
					Message: "compound",
				}}
			}
		}
		return nil
	})
	matches := lt.Check("Stundenkilometer")
	require.Len(t, matches, 1)
	require.Equal(t, 0, matches[0].FromPos)
	require.Equal(t, len([]rune("Stundenkilometer")), matches[0].ToPos)
}

// Port of JLanguageToolTest.testCleanOverlappingWithGerman
func TestJLanguageTool_lang_de_CleanOverlappingWithGerman(t *testing.T) {
	// soft: three non-overlapping synthetic matches survive cleaning
	matches := []LocalMatch{
		{FromPos: 0, ToPos: 4, RuleID: "A", Priority: 1},
		{FromPos: 5, ToPos: 6, RuleID: "B", Priority: 1},
		{FromPos: 7, ToPos: 17, RuleID: "C", Priority: 1},
	}
	// TRGS - Technische… style adjacent errors should not remove each other
	cleaned := CleanOverlappingLocalMatches(matches)
	require.Len(t, cleaned, 3)

	// overlapping: higher priority wins
	ov := []LocalMatch{
		{FromPos: 0, ToPos: 5, RuleID: "low", Priority: 1},
		{FromPos: 2, ToPos: 8, RuleID: "high", Priority: 10},
	}
	cleaned = CleanOverlappingLocalMatches(ov)
	require.Len(t, cleaned, 1)
	require.Equal(t, "high", cleaned[0].RuleID)
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
		lt.AddChecker(SimpleWordRepeatChecker(""))
		require.Equal(t, code, lt.GetLanguageCode())
		require.Empty(t, lt.Check("Hallo Welt."))
		require.NotEmpty(t, lt.Check("Hallo Hallo Welt."))
	}
}

// Port of JLanguageToolTest.testMultitokenSpeller
func TestJLanguageTool_lang_de_MultitokenSpeller(t *testing.T) {
	lt := NewJLanguageTool("de-DE")
	// soft: multi-token analyze; speller inject accepts phrase tokens separately
	lt.IsKnownWord = KnownWordSet("New", "York", "City")
	lt.SetListUnknownWords(true)
	_ = lt.Check("New York City")
	require.Empty(t, lt.GetUnknownWords())
}
