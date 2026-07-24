package spelling

// Twin of VagueSpellCheckerTest.testIsValidWord
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVagueSpellChecker_IsValidWord(t *testing.T) {
	checker := NewVagueSpellChecker()
	// Inject per-language validators (full Morfologik/Hunspell deferred)
	en := map[string]bool{"vacation": true, "walks": true}
	de := map[string]bool{"Hütte": true, "Hütten": true}
	pt := map[string]bool{"termo": true}
	fr := map[string]bool{"voiture": true}
	checker.Register("en", func(w string) bool { return en[w] })
	checker.Register("en-US", func(w string) bool { return en[w] })
	checker.Register("de", func(w string) bool { return de[w] })
	checker.Register("de-DE", func(w string) bool { return de[w] })
	checker.Register("pt", func(w string) bool { return pt[w] })
	checker.Register("pt-PT", func(w string) bool { return pt[w] })
	checker.Register("fr", func(w string) bool { return fr[w] })

	require.True(t, checker.IsValidWord("vacation", "en-US"))
	require.True(t, checker.IsValidWord("walks", "en-US"))
	require.False(t, checker.IsValidWord("vacationx", "en-US"))

	require.True(t, checker.IsValidWord("Hütte", "de-DE"))
	require.True(t, checker.IsValidWord("Hütten", "de-DE"))
	require.False(t, checker.IsValidWord("sdasfd", "de-DE"))

	require.True(t, checker.IsValidWord("termo", "pt-PT"))
	require.False(t, checker.IsValidWord("termoasq", "pt-PT"))
	require.False(t, checker.IsValidWord("difdsf", "pt-PT"))

	require.True(t, checker.IsValidWord("voiture", "fr"))
	require.False(t, checker.IsValidWord("sduiofhdf", "fr"))
}
