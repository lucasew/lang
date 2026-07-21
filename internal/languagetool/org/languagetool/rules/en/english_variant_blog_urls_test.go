package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnVariantBlogURL_FullMatchJava(t *testing.T) {
	// Java Matcher.matches(): archeological is pattern[8], not substring .*og
	require.Equal(t,
		"https://languagetool.org/insights/post/our-or/#likeable-vs-likable-judgement-vs-judgment-oestrogen-vs-estrogen",
		enVariantBlogURL("archeological"))
	// colour matches *(or|our)* full string
	require.Contains(t, enVariantBlogURL("colour"), "colour-or-color")
	// catalog ends with og → quillbot catalog/og pattern
	require.Equal(t, "https://quillbot.com/blog/category/uk-vs-us/", enVariantBlogURL("catalog"))
	require.Equal(t, "", enVariantBlogURL(""))
	require.Equal(t, "", enVariantBlogURL("hello"))
}
