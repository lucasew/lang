package languagetool

// Twin of AnnotatedTextBuilderTest — markup surface + CheckAnnotated inject.
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
	require.Greater(t, len([]rune(text.GetTextWithMarkup())), len([]rune(text.GetPlainText())))
}

// Port of AnnotatedTextBuilderTest.testWithEmptyFakeContent
func TestAnnotatedTextBuilder_WithEmptyFakeContent(t *testing.T) {
	text := markup.NewAnnotatedTextBuilder().
		AddText("And ths is ").
		AddMarkupInterpretAs("_", "").
		Build()
	require.Equal(t, "And ths is ", text.GetPlainText())
	require.Equal(t, "And ths is _", text.GetTextWithMarkup())

	// CheckAnnotated with speller inject on typo "ths"
	lt := NewJLanguageTool("en")
	known := map[string]struct{}{"And": {}, "is": {}, "this": {}}
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", known, map[string][]string{
		"ths": {"this"},
	}))
	m := lt.CheckAnnotated(text)
	require.NotEmpty(t, m)
	fixed := CorrectTextFromLocalMatches(text.GetPlainText(), m)
	require.Equal(t, "And this is ", fixed)
}
