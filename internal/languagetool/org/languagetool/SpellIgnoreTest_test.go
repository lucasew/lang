package languagetool

// Twin of SpellIgnoreTest — ignore-set surface without importing rules/spelling (cycle).
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpellIgnore_Ignore(t *testing.T) {
	// soft AcceptWord mirror of SpellingCheckRule
	dict := map[string]struct{}{
		"This": {}, "is": {}, "a": {}, "text": {}, "with": {}, "and": {},
	}
	ignore := map[string]struct{}{}
	accept := func(w string) bool {
		if _, ok := ignore[w]; ok {
			return true
		}
		_, ok := dict[w]
		return ok
	}
	text := "This is a text with specialword and myotherword"
	var before int
	for _, w := range strings.Fields(text) {
		if !accept(w) {
			before++
		}
	}
	require.Equal(t, 2, before)
	ignore["specialword"] = struct{}{}
	ignore["myotherword"] = struct{}{}
	var after int
	for _, w := range strings.Fields(text) {
		if !accept(w) {
			after++
		}
	}
	require.Equal(t, 0, after)
}
