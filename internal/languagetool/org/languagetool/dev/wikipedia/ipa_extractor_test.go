package wikipedia

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIpaExtractor_FullPattern(t *testing.T) {
	e := NewIpaExtractor()
	n := e.ExtractFromText("Trance", "'''Trance''' [{{IPA|trɑ̃s}}] is music.")
	require.Equal(t, 1, n)
	require.Equal(t, 1, e.IPACount)
	require.Equal(t, "Trance", e.Hits[0].Word)
	require.Equal(t, "trɑ̃s", e.Hits[0].IPA)
}

func TestIpaExtractor_BareIPA(t *testing.T) {
	e := NewIpaExtractor()
	n := e.ExtractFromText("X", "something {{IPA|ˈfoo}} more")
	require.Equal(t, 1, n)
	require.Equal(t, "ˈfoo", e.Hits[0].IPA)
	require.Empty(t, e.Hits[0].Word)
}

func TestIpaExtractor_NoIPA(t *testing.T) {
	e := NewIpaExtractor()
	require.Equal(t, 0, e.ExtractFromText("Y", "no phonetics here"))
	require.Equal(t, 0, e.IPACount)
}

func TestIpaExtractor_FromXML(t *testing.T) {
	xml := `<mediawiki>
	<page><title>Trance</title><revision><text>'''Trance''' [{{IPA|trɑ̃s}}]</text></revision></page>
	<page><title>Other</title><revision><text>no ipa</text></revision></page>
</mediawiki>`
	e := NewIpaExtractor()
	require.NoError(t, e.ExtractFromMediaWikiXML(strings.NewReader(xml)))
	require.Equal(t, 2, e.ArticleCount)
	require.Equal(t, 1, e.IPACount)
	require.Equal(t, "Trance", e.Hits[0].Title)
}
