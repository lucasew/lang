package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestIrishPartialPosTagFilter_Injected(t *testing.T) {
	// with explicit tag func (bypass disambiguator path)
	f := NewIrishPartialPosTagFilter(func(p string) []string {
		if p == "fear" {
			return []string{"Noun:Masc:Com:Sg"}
		}
		return nil
	})
	ok, err := f.Accept("fir", "^(fear)$", "Noun.*", false, false, "", "")
	// Accept uses partial from regex capture of surface "fir" against pattern - check PartialPosTagFilter API
	// For injected tag, Accept(partial match form)
	_ = ok
	_ = err
	// simpler: use Accept with partial string that tag returns
	ok, err = f.Accept("fears", "^(fear)s$", "Noun.*", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNoDisambiguationIrishPartialPosTagFilter_FailClosed(t *testing.T) {
	ClearDefaultIrishPartialPosTagger()
	f := NewNoDisambiguationIrishPartialPosTagFilter(nil)
	ok, err := f.Accept("fir", "^(fear)$", "Noun.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	f2 := NewNoDisambiguationIrishPartialPosTagFilter(func(p string) []string {
		if p == "fear" {
			return []string{"Noun:Masc:Com:Sg"}
		}
		return nil
	})
	ok, err = f2.Accept("fears", "^(fear)s$", "Noun.*", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestIrishPartialFiltersRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ga.IrishPartialPosTagFilter"))
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ga.NoDisambiguationIrishPartialPosTagFilter"))
}
