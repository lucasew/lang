package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalanAdaptSuggestion_GensAndContractions(t *testing.T) {
	// gens traça / facilitat inserts
	require.Equal(t, "gens de traça", CatalanAdaptSuggestion("gens traça", ""))
	require.Equal(t, "gens de facilitat", CatalanAdaptSuggestion("gens facilitat", ""))
	// CA_CONTRACTIONS: a el → al, de el → del, per el → pel, etc.
	require.Equal(t, "al llibre", CatalanAdaptSuggestion("a el llibre", ""))
	require.Equal(t, "del mar", CatalanAdaptSuggestion("de el mar", ""))
	require.Equal(t, "pels dies", CatalanAdaptSuggestion("per els dies", ""))
}

func TestCatalanAdaptSuggestion_Apostrophes1(t *testing.T) {
	// CA_APOSTROPHES1: "l' " → "l'"
	require.Equal(t, "l'home", CatalanAdaptSuggestion("l' home", ""))
	require.Equal(t, "d'aigua", CatalanAdaptSuggestion("d' aigua", ""))
}

func TestCatalanAdaptSuggestion_Apostrophes8(t *testing.T) {
	// T'comença → Et comença
	require.Equal(t, "Et comença", CatalanAdaptSuggestion("T'comença", ""))
	require.Equal(t, "Es veu", CatalanAdaptSuggestion("S'veu", "S'veu"))
}

func TestCatalanAdaptSuggestion_RemoveSpacesAndComma(t *testing.T) {
	// CA_REMOVE_SPACES: a ls → als (when not followed by apostrophe)
	require.Equal(t, "als homes", CatalanAdaptSuggestion("a ls homes", ""))
	require.Equal(t, "hola, món", CatalanAdaptSuggestion("hola , món", ""))
}

func TestCatalanAdaptSuggestion_PreserveCase(t *testing.T) {
	// Capitalized input re-uppercases first char after transforms
	got := CatalanAdaptSuggestion("L' home", "L' home")
	require.Equal(t, "L'home", got)
}

func TestCatalanAdaptSuggestion_UnicodeWordBoundary(t *testing.T) {
	// Java UNICODE_CHARACTER_CLASS \b: letter à is a word char, so "xà a el y" must not
	// contract across à|space|a (left edge of "a el" fails because previous is space
	// and after space "a" starts a word — actually left bound of "a el" is after space: OK.
	// Mid-word: "xael" should not match contractions.
	require.Equal(t, "xael", CatalanAdaptSuggestion("xael", ""))
	// Non-ASCII letter before candidate must block left \b: "ça a el" — left of "a el" is space → OK → "çal"
	// Wait: "ça a el" → contraction "a el" → "al" → "ça al"? 
	// "ça " + "a el" → left bound of "a el" is true (space), right after "el" true → "ça al"
	require.Equal(t, "ça al mar", CatalanAdaptSuggestion("ça a el mar", ""))
	// Letter glued without boundary: "çaa el" should not become "çaal" via "a el"
	// Pattern is "a el" with space; "çaa el" has "a el" with left bound after second 'a'? 
	// "çaa el": match "a el" starting at second a: left prev is 'a' word char → blocked
	require.Equal(t, "çaa el", CatalanAdaptSuggestion("çaa el", ""))
}
