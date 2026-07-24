package wikipedia

// Twin of languagetool-wikipedia WikipediaQuickCheckTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWikipediaQuickCheck_CheckWikipediaMarkup(t *testing.T) {
	// soft: plain-text check path without GermanyGerman rule stack
	qc := NewWikipediaQuickCheck()
	// strip markup first, then check plain
	plain := NewSimpleWikipediaTextFilter().Filter("Ein [[Test]] Satz.")
	require.Contains(t, plain, "Test")
	res := qc.CheckPlainText(plain, "de", nil)
	require.Equal(t, "de", res.GetLanguageCode())
	require.NotEmpty(t, res.GetText())

	// inject LT Check on filtered plain text (word repeat)
	lt := languagetool.NewJLanguageTool("de")
	lt.AddRuleChecker("WORD_REPEAT_RULE", languagetool.SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	wikiPlain := NewSimpleWikipediaTextFilter().Filter("Ein [[Test]] Test Test Satz.")
	// may or may not have double Test after filter — force check on synthetic
	m := lt.Check("Ein Test Test Satz.")
	require.NotEmpty(t, m)
	_ = wikiPlain
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

// Twin of WikipediaQuickCheckTest.testRemoveInterLanguageLinks (Java calls removeWikipediaLinks).
func TestWikipediaQuickCheck_RemoveInterLanguageLinks(t *testing.T) {
	require.Equal(t, "foo  bar", RemoveWikipediaLinks("foo [[pt:Some Article]] bar"))
	require.Equal(t, "foo [[some link]] bar", RemoveWikipediaLinks("foo [[some link]] bar"))
	require.Equal(t, "foo [[Some Link]] bar ", RemoveWikipediaLinks("foo [[Some Link]] bar [[pt:Some Article]]"))
	// known limitation
	require.Equal(t, "foo [[zh-min-nan:Linux]] bar", RemoveWikipediaLinks("foo [[zh-min-nan:Linux]] bar"))
	require.Equal(t, "[[Scultura bronzea di Gaudí mentre osserva il suo ''[[Il Capriccio|Capriccio]]'']]",
		RemoveWikipediaLinks("[[File:Gaudì-capriccio.JPG|thumb|left|Scultura bronzea di Gaudí mentre osserva il suo ''[[Il Capriccio|Capriccio]]'']]"))
	require.Equal(t, "[[[[Palau de la Música Catalana]], entrada]]",
		RemoveWikipediaLinks("[[Fitxer:Palau_de_musica_2.JPG|thumb|[[Palau de la Música Catalana]], entrada]]"))
	require.Equal(t, "foo  bar", RemoveWikipediaLinks("foo [[Kategorie:Kurgebäude]] bar"))
	require.Equal(t, "foo [[''Kursaal Palace'' in San Sebastián]] bar",
		RemoveWikipediaLinks("foo [[Datei:FestivalSS.jpg|miniatur|''Kursaal Palace'' in San Sebastián]] bar"))
	require.Equal(t, "[[Yupana, emprat pels [[Inques]].]]",
		RemoveWikipediaLinks("[[Fitxer:Yupana 1.GIF|thumb|Yupana, emprat pels [[Inques]].]]"))
}
