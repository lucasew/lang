package spelling

// Twin of SpellIgnoreTest (surface) — full EN Morfologik deferred.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpellIgnore_Ignore(t *testing.T) {
	text := "This is a text with specialword and myotherword"
	dict := map[string]struct{}{
		"This": {}, "is": {}, "a": {}, "text": {}, "with": {}, "and": {},
	}
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN_US", "spelling", "en")
	r.IsMisspelled = func(word string) bool {
		_, ok := dict[word]
		return !ok
	}
	var before int
	for _, w := range strings.Fields(text) {
		if !r.AcceptWord(w) {
			before++
		}
	}
	require.Equal(t, 2, before)
	r.AddIgnoreWords("specialword", "myotherword")
	var after int
	for _, w := range strings.Fields(text) {
		if !r.AcceptWord(w) {
			after++
		}
	}
	require.Equal(t, 0, after)
}
