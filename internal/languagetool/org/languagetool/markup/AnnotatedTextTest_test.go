package markup

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.markup.AnnotatedTextTest.

func TestAnnotatedText_Test(t *testing.T) {
	text := NewAnnotatedTextBuilder().
		AddGlobalMetaData("foo", "val").
		AddGlobalMetaDataKey(MetaEmailToAddress, "Foo Bar <foo@foobar.org>").
		AddText("hello ").
		AddMarkup("<b>").
		AddText("user!").
		AddMarkup("</b>").
		Build()

	require.Equal(t, "val", text.GetGlobalMetaDataString("foo", ""))
	require.Equal(t, "xxx", text.GetGlobalMetaDataString("non-existing-key", "xxx"))
	require.Equal(t, "Foo Bar <foo@foobar.org>", text.GetGlobalMetaDataKey(MetaEmailToAddress, "xxx"))
	require.Equal(t, "default-title", text.GetGlobalMetaDataKey(MetaDocumentTitle, "default-title"))
	require.Equal(t, "hello user!", text.GetPlainText())
	require.Equal(t, 0, text.GetOriginalTextPositionFor(0, false))
	require.Equal(t, 5, text.GetOriginalTextPositionFor(5, false))
	require.Equal(t, 9, text.GetOriginalTextPositionFor(6, false))
	require.Equal(t, 10, text.GetOriginalTextPositionFor(7, false))
	// hello user!  ^ = pos 8 → hello <b>user!</b> ^ = pos 11
	require.Equal(t, 11, text.GetOriginalTextPositionFor(8, false))
}

func TestAnnotatedText_IgnoreInterpretAs(t *testing.T) {
	text := NewAnnotatedTextBuilder().
		AddText("hello ").
		AddMarkupInterpretAs("<p>", "\n\n").
		AddText("more xxxx text!").
		Build()
	require.Equal(t, "hello \n\nmore xxxx text!", text.GetPlainText())
	require.Equal(t, "hello <p>more xxxx text!", text.GetTextWithMarkup())
	ct := tools.NewContextTools()
	ct.SetErrorMarker("#", "#")
	ct.SetEscapeHtml(false)
	require.Equal(t, "hello <p>more #xxxx# text!", ct.GetContext(14, 18, text.GetTextWithMarkup()))
}
