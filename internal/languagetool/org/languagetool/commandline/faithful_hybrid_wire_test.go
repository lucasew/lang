package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Official multiwords + disambiguation.xml must resolve (submodule or vendored).
func TestDiscoverOfficialDisambigResources(t *testing.T) {
	mw := DiscoverEnglishMultiwords(nil)
	require.NotEmpty(t, mw, "en multiwords.txt")
	require.NotContains(t, mw, "soft.txt", "must not use soft multiword invent list")

	xml := DiscoverLanguageDisambiguationXML(nil, "en")
	require.NotEmpty(t, xml, "en disambiguation.xml")
	require.NotContains(t, xml, "soft.xml")
	require.Contains(t, xml, "disambiguation.xml")

	// global optional but preferred when present
	_ = DiscoverGlobalDisambiguationXML(nil)
	_ = DiscoverSpellingGlobal(nil)
}

func TestRegisterEnglishHybridDisambiguator(t *testing.T) {
	lt, err := configureCoreLT("en", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	require.NotNil(t, lt)
	// When official resources exist, hybrid is installed.
	if DiscoverEnglishMultiwords(nil) != "" || DiscoverLanguageDisambiguationXML(nil, "en") != "" {
		require.NotNil(t, lt.Disambiguator, "EnglishHybridDisambiguator should be wired")
	}
	// Smoke: analyze does not panic
	sents := lt.Analyze("New York is big.")
	require.NotEmpty(t, sents)
}
