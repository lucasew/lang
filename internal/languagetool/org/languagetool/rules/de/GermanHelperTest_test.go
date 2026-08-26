package de

// Twin of org.languagetool.rules.de.GermanHelperTest (+ method surface not covered there).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// testHasReadingOfType twin of GermanHelperTest.testHasReadingOfType
func TestGermanHelper_HasReadingOfType(t *testing.T) {
	// Java: new AnalyzedTokenReadings(new AnalyzedToken("der", "ART:DEF:DAT:SIN:FEM", null), 0)
	tag := "ART:DEF:DAT:SIN:FEM"
	tok := languagetool.NewAnalyzedTokenStr("der", tag, "", false, true)
	readings := languagetool.NewAnalyzedTokenReadings(tok)
	require.True(t, HasReadingOfType(readings, POSDeterminer))
	require.False(t, HasReadingOfType(readings, POSNomen))
}

// testGetDeterminerNumber twin
func TestGermanHelper_GetDeterminerNumber(t *testing.T) {
	require.Equal(t, "SIN", GetDeterminerNumber("ART:DEF:DAT:SIN:FEM"))
}

// testGetDeterminerDefiniteness twin
func TestGermanHelper_GetDeterminerDefiniteness(t *testing.T) {
	require.Equal(t, "DEF", GetDeterminerDefiniteness("ART:DEF:DAT:SIN:FEM"))
}

// testGetDeterminerCase twin
func TestGermanHelper_GetDeterminerCase(t *testing.T) {
	require.Equal(t, "DAT", GetDeterminerCase("ART:DEF:DAT:SIN:FEM"))
}

// testGetDeterminerGender twin (Java null → "" maps to Go empty string)
func TestGermanHelper_GetDeterminerGender(t *testing.T) {
	// Java: GermanHelper.getDeterminerGender(null) is ""
	require.Equal(t, "", GetDeterminerGender(""))
	require.Equal(t, "", GetDeterminerGender(""))
	require.Equal(t, "FEM", GetDeterminerGender("ART:DEF:DAT:SIN:FEM"))
}

// Extra surface: getNoun* (Java methods; not in GermanHelperTest)
func TestGermanHelper_GetNounFields(t *testing.T) {
	// input e.g. SUB:AKK:SIN:NEU
	require.Equal(t, "AKK", GetNounCase("SUB:AKK:SIN:NEU"))
	require.Equal(t, "SIN", GetNounNumber("SUB:AKK:SIN:NEU"))
	require.Equal(t, "NEU", GetNounGender("SUB:AKK:SIN:NEU"))
	require.Equal(t, "", GetNounCase(""))
	require.Equal(t, "", GetNounNumber("SUB"))
}

// Extra surface: getComparison
func TestGermanHelper_GetComparison(t *testing.T) {
	// ADJ:AKK:PLU:MAS:KOM:SOL → index 4 = KOM
	require.Equal(t, "KOM", GetComparison("ADJ:AKK:PLU:MAS:KOM:SOL"))
	require.Equal(t, "GRU", GetComparison("ADJ:NOM:SIN:NEU:GRU:SOL"))
	require.Equal(t, "SUP", GetComparison("ADJ:NOM:SIN:NEU:SUP:SOL"))
	// PA2:PRD:GRU:VER — index 4 empty, fall back to index 2
	require.Equal(t, "GRU", GetComparison("PA2:PRD:GRU:VER"))
	require.Equal(t, "", GetComparison(""))
	require.Equal(t, "", GetComparison("ADJ"))
}

// HasReadingOfType null / SENTENCE_END / PARAGRAPH_END control flow (Java GermanHelper)
func TestGermanHelper_HasReadingOfType_EdgeCases(t *testing.T) {
	require.False(t, HasReadingOfType(nil, POSDeterminer))

	sent := languagetool.SentenceEndTagName
	sentTok := languagetool.NewAnalyzedTokenStr(".", sent, "", false, true)
	require.False(t, HasReadingOfType(languagetool.NewAnalyzedTokenReadings(sentTok), POSDeterminer))

	para := languagetool.ParagraphEndTagName
	paraTok := languagetool.NewAnalyzedTokenStr("\n", para, "", false, true)
	require.False(t, HasReadingOfType(languagetool.NewAnalyzedTokenReadings(paraTok), POSNomen))

	// null POS tag → AnalyzedGermanToken type null → no match
	nullPOS := languagetool.NewAnalyzedTokenStr("x", "", "", true, true)
	require.False(t, HasReadingOfType(languagetool.NewAnalyzedTokenReadings(nullPOS), POSNomen))

	// short tag (<3 parts) → type null
	adv := "ADV"
	advTok := languagetool.NewAnalyzedTokenStr("schnell", adv, "", false, true)
	require.False(t, HasReadingOfType(languagetool.NewAnalyzedTokenReadings(advTok), POSAdjective))

	// NOMEN / VERB / ADJEKTIV / PROPER_NOUN / PARTIZIP via ParseGermanPOS path
	cases := []struct {
		tag  string
		want POSType
		no   POSType
	}{
		{"SUB:NOM:SIN:NEU", POSNomen, POSVerb},
		{"EIG:NOM:SIN:MAS", POSProperNoun, POSNomen},
		{"VER:PA2:SFT", POSPartizip, POSVerb}, // PA2 overrides; length ≥ 3
		{"ADJ:NOM:SIN:NEU:GRU:SOL", POSAdjective, POSNomen},
		{"PRO:PER:NOM:SIN:MAS", POSPronoun, POSDeterminer},
		{"VER:1:SIN:PRÄ:SFT", POSVerb, POSPartizip},
	}
	for _, c := range cases {
		tok := languagetool.NewAnalyzedTokenStr("w", c.tag, "", false, true)
		r := languagetool.NewAnalyzedTokenReadings(tok)
		require.True(t, HasReadingOfType(r, c.want), "tag %s want %s", c.tag, c.want)
		require.False(t, HasReadingOfType(r, c.no), "tag %s not %s", c.tag, c.no)
	}
}

// getIndexOrEmptyString null/empty/short-tag behavior
func TestGermanHelper_GetIndexOrEmptyString(t *testing.T) {
	require.Equal(t, "", getIndexOrEmptyString("", 0))
	require.Equal(t, "", getIndexOrEmptyString("", 1))
	require.Equal(t, "ART", getIndexOrEmptyString("ART:DEF:DAT:SIN:FEM", 0))
	require.Equal(t, "FEM", getIndexOrEmptyString("ART:DEF:DAT:SIN:FEM", 4))
	require.Equal(t, "", getIndexOrEmptyString("ART:DEF", 4))
	require.Equal(t, "", getIndexOrEmptyString("ART", 1))
}
