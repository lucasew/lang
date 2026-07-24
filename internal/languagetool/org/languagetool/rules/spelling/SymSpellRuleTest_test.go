package spelling

// Twin of SymSpellRuleTest (Java has no @Test) — surface smoke if SymSpellRule exists.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSymSpellRule_NoTests(t *testing.T) {
	// package may define SymSpell helpers; ensure SpellingCheckRule still works as fallback
	r := NewSpellingCheckRule("SYMSPELL", "sym", "en")
	require.Equal(t, "SYMSPELL", r.GetID())
	require.NotEmpty(t, r.GetDescription())
}
