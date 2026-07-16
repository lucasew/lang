package languagetool

// Twin of AnnotatedTextBuilderTest — full EN grammar check deferred; markup plain-text surface.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/stretchr/testify/require"
)

// Port of AnnotatedTextBuilderTest.test
func TestAnnotatedTextBuilder_Test(t *testing.T) {
	text := markup.NewAnnotatedTextBuilder().
		AddText("This is a caf").
		AddMarkupInterpretAs("&eacute;", "é").
		Build()
	require.Equal(t, "This is a café", text.GetPlainText())
	require.Equal(t, "This is a caf&eacute;", text.GetTextWithMarkup())
	// markup length of &eacute; is greater than interpretAs é
	require.Greater(t, len([]rune(text.GetTextWithMarkup())), len([]rune(text.GetPlainText())))
}

// Port of AnnotatedTextBuilderTest.testWithEmptyFakeContent
func TestAnnotatedTextBuilder_WithEmptyFakeContent(t *testing.T) {
	text := markup.NewAnnotatedTextBuilder().
		AddText("And ths is ").
		AddMarkupInterpretAs("_", "").
		Build()
	// empty interpretAs → plain text is just the real text (typo "ths" preserved)
	require.Equal(t, "And ths is ", text.GetPlainText())
	require.Equal(t, "And ths is _", text.GetTextWithMarkup())
}
