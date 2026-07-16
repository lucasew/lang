package hunspell

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestCompoundAwareHunspellSuggest(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"well", "known", "wellknown"})
	morfo := morfologik.NewMorfologikSpeller("en", 1)
	morfo.AddWord("well")
	morfo.AddWord("known")
	morfo.Suggestions["wel"] = []string{"well"}
	multi := morfologik.NewMorfologikMultiSpeller(morfo)
	r := NewCompoundAwareHunspellRule("en", dict, nil, multi)
	// compound split of well-known
	sug := r.Suggest("well-known")
	require.NotEmpty(t, sug)
	// misspelled with morfo suggestion
	dict2 := NewMapHunspellDictionary([]string{"well", "known"})
	r2 := NewCompoundAwareHunspellRule("en", dict2, nil, multi)
	require.NotEmpty(t, r2.Suggest("wel-known"))
}
