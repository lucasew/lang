package spelling

// Twin of SpellcheckerTest (Java has no @Test) — SpellingCheckRule smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpellchecker_NoTests(t *testing.T) {
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN", "spell", "en")
	r.IsMisspelled = func(w string) bool { return w == "xyzzy" }
	require.True(t, r.AcceptWord("hello"))
	require.False(t, r.AcceptWord("xyzzy"))
	r.AddIgnoreWords("xyzzy")
	require.True(t, r.AcceptWord("xyzzy"))
}
