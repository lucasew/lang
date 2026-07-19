package nl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWireDutchFilterTagger_FailClosed(t *testing.T) {
	ClearDutchFilterTagger()
	t.Cleanup(ClearDutchFilterTagger)
	require.False(t, WireDutchFilterTagger(""))
	require.False(t, WireDutchFilterTagger("/no/such/dutch.dict"))
	require.False(t, FilterTaggerAvailable())
	require.Nil(t, FilterGetPostags("puzzel"))
}

func TestTryWireDutchFilterTagger_OptionalDict(t *testing.T) {
	ClearDutchFilterTagger()
	t.Cleanup(ClearDutchFilterTagger)
	// dutch.dict is often not vendored; skip when missing (no invent).
	if !TryWireDutchFilterTagger() {
		t.Skip("dutch.dict not discoverable")
	}
	require.True(t, FilterTaggerAvailable())
	require.Nil(t, FilterGetPostags(""))
	BindDefaultCompoundAcceptorFilters()
	require.NotNil(t, DefaultCompoundAcceptor.TagPOS)
}

func TestBindDefaultCompoundAcceptorFilters_NoTagger(t *testing.T) {
	ClearDutchFilterTagger()
	prev := DefaultCompoundAcceptor.TagPOS
	t.Cleanup(func() {
		DefaultCompoundAcceptor.TagPOS = prev
		ClearDutchFilterTagger()
	})
	DefaultCompoundAcceptor.TagPOS = nil
	BindDefaultCompoundAcceptorFilters()
	// without tagger, TagPOS stays nil (fail-closed)
	require.Nil(t, DefaultCompoundAcceptor.TagPOS)
}

func TestCompoundAcceptor_TagPOS_ZNW(t *testing.T) {
	c := NewCompoundAcceptor()
	require.NoError(t, c.LoadNoS(strings.NewReader("straat\n")))
	c.TagPOS = func(word string) []string {
		if word == "puzzel" {
			return []string{"ZNW:EKV:DE_"}
		}
		return nil
	}
	c.SpellingOk = func(w string) bool { return w == "straat" || w == "puzzel" }
	require.True(t, c.AcceptCompoundParts("straat", "puzzel"))
	// part2Exceptions block noun
	c.Part2Exceptions["puzzel"] = struct{}{}
	require.False(t, c.AcceptCompoundParts("straat", "puzzel"))
}

func TestTryWireDutchFilterSpeller_OptionalDict(t *testing.T) {
	ClearDutchFilterSpeller()
	t.Cleanup(ClearDutchFilterSpeller)
	if !TryWireDutchFilterSpeller() {
		t.Skip("nl_NL.dict not discoverable")
	}
	require.True(t, FilterDictAvailable())
	// nonsense is misspelled when dict is live
	require.True(t, FilterDictIsMisspelled("xyzzyqqqnotaword"))
}
