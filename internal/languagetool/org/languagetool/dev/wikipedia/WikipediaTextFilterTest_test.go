package wikipedia

// Twin of WikipediaTextFilterTest (Java @Ignore class; green simple filter)
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func assertExtract(t *testing.T, input, expected string) {
	t.Helper()
	got := NewSimpleWikipediaTextFilter().Filter(input)
	require.Equal(t, expected, got)
}

func TestWikipediaTextFilter_ImageRemoval(t *testing.T) {
	assertExtract(t,
		"foo [[Datei:Bundesarchiv Bild 183-1990-0803-017.jpg|miniatur|Mit Lothar de Maizière im August 1990]] bar",
		"foo bar")
}

func TestWikipediaTextFilter_RemovalOfImageWithLink(t *testing.T) {
	assertExtract(t,
		"foo [[Datei:Bundesarchiv Bild 183-1990-0803-017.jpg|miniatur|Mit [[Lothar de Maizière]] im August 1990]] bar [[Link]]",
		"foo bar Link")
}

func TestWikipediaTextFilter_Link1(t *testing.T) {
	assertExtract(t, "foo [[Test]] bar", "foo Test bar")
}

func TestWikipediaTextFilter_Link2(t *testing.T) {
	assertExtract(t, "foo [[Target|visible link]] bar", "foo visible link bar")
}

func TestWikipediaTextFilter_Entity(t *testing.T) {
	assertExtract(t, "rund 20&nbsp;Kilometer südlich", "rund 20\u00A0Kilometer südlich")
	assertExtract(t, "one&lt;br/&gt;two", "one<br/>two")
	assertExtract(t, "one &ndash; two", "one – two")
	assertExtract(t, "one &mdash; two", "one — two")
	assertExtract(t, "one &amp; two", "one & two")
}

func TestWikipediaTextFilter_Lists(t *testing.T) {
	assertExtract(t, "# one\n# two\n", "one\n\ntwo")
	assertExtract(t, "* one\n* two\n", "one\n\ntwo")
}

func TestWikipediaTextFilter_OtherStuff(t *testing.T) {
	assertExtract(t,
		"Daniel Guerin, ''[http://theanarchistlibrary.org Anarchism: From Theory to Practice]''",
		"Daniel Guerin, Anarchism: From Theory to Practice")
	assertExtract(t,
		"foo <ref>\"At the end of the century in France [http://theanarchistlibrary.org] [[Daniel Guérin]]. ''Anarchism'']</ref>",
		"foo")
	assertExtract(t, "The <code>$pattern</code>", "The $pattern")
	assertExtract(t, "foo <source lang=\"bash\">some source</source> bar", "foo bar")
}
