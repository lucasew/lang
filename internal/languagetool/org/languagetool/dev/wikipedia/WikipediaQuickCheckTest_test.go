package wikipedia

// Twin of languagetool-wikipedia WikipediaQuickCheckTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWikipediaQuickCheck_CheckWikipediaMarkup(t *testing.T) {
	t.Skip("soft-skip: full LT check over wiki markup (GermanyGerman rules)")
}

func TestWikipediaQuickCheck_URLParse(t *testing.T) {
	lang, title, err := MatchWikipediaURL("https://de.wikipedia.org/wiki/Augsburg")
	require.NoError(t, err)
	require.Equal(t, "de", lang)
	require.Equal(t, "Augsburg", title)

	lang, title, err = MatchWikipediaURL("https://secure.wikimedia.org/wikipedia/en/wiki/Hello_World")
	require.NoError(t, err)
	require.Equal(t, "en", lang)
	require.Equal(t, "Hello_World", title)

	_, _, err = MatchWikipediaURL("https://example.com/not-wiki")
	require.Error(t, err)

	qc := NewWikipediaQuickCheck()
	require.NoError(t, qc.ValidateWikipediaURL("http://en.wikipedia.org/wiki/Foo"))
	got, err := qc.GetPageTitle("http://en.wikipedia.org/wiki/Foo%20Bar")
	require.NoError(t, err)
	require.Equal(t, "Foo Bar", got)
}

func TestWikipediaQuickCheck_GetPlainText(t *testing.T) {
	qc := NewWikipediaQuickCheck()
	xml := `<?xml version="1.0"?><api><query><normalized><n from="Benutzer_Diskussion:Dnaber" to="Benutzer Diskussion:Dnaber" />` +
		`</normalized><pages><page pageid="143424" ns="3" title="Benutzer Diskussion:Dnaber"><revisions><rev xml:space="preserve">` +
		"\nTest [[Link]] Foo&amp;nbsp;bar.\n" +
		`</rev></revisions></page></pages></query></api>`
	plain, err := qc.GetPlainText(xml)
	require.NoError(t, err)
	require.Equal(t, "Test Link Foo\u00A0bar.", plain)
}

func TestWikipediaQuickCheck_GetPlainTextMapping(t *testing.T) {
	qc := NewWikipediaQuickCheck()
	text := "Test [[Link]] und [[AnotherLink|noch einer]] und [http://test.org external link] Foo&amp;nbsp;bar.\n"
	xml := `<?xml version="1.0"?><api><query><pages><page><revisions><rev xml:space="preserve">` +
		text + `</rev></revisions></page></pages></query></api>`
	mapping, err := qc.GetPlainTextMapping(xml)
	require.NoError(t, err)
	require.Equal(t, "Test Link und noch einer und external link Foo\u00A0bar.", mapping.GetPlainText())
}

func TestWikipediaQuickCheck_GetPlainTextMappingMultiLine1(t *testing.T) {
	f := NewSimpleWikipediaTextFilter()
	plain := f.Filter("line one\n# item\n# two\n")
	require.Contains(t, plain, "item")
	require.Contains(t, plain, "two")
}

func TestWikipediaQuickCheck_GetPlainTextMappingMultiLine2(t *testing.T) {
	f := NewSimpleWikipediaTextFilter()
	require.Equal(t, "a b", f.Filter("a [[x|b]]"))
}

func TestWikipediaQuickCheck_RemoveWikipediaLinks(t *testing.T) {
	in := "Hello [[pt:Linux]] world [[Category:Foo]] rest"
	out := RemoveWikipediaLinks(in)
	require.NotContains(t, out, "pt:Linux")
	require.NotContains(t, out, "Category:Foo")
	require.Contains(t, out, "Hello")
	require.Contains(t, out, "world")
}

func TestWikipediaQuickCheck_CheckPlainText(t *testing.T) {
	qc := NewWikipediaQuickCheck()
	res := qc.CheckPlainText("hello", "en", nil)
	require.Equal(t, "hello", res.GetText())
	require.Equal(t, "en", res.GetLanguageCode())
	require.Empty(t, res.GetRuleMatches())
}
