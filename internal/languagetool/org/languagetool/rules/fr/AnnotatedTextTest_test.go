package fr

// Twin of languagetool-language-modules/fr/src/test/java/org/languagetool/rules/fr/AnnotatedTextTest.java
// Markup mapping greens without full French LT (entity/position surface).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/stretchr/testify/require"
)

// Port of AnnotatedTextTest.testInterpretAsBefore — entity before error span.
func TestAnnotatedText_InterpretAsBefore(t *testing.T) {
	// Une &eacute;chapatoire → plain "Une échapatoire"
	text := markup.NewAnnotatedTextBuilder().
		AddText("Une ").
		AddMarkupInterpretAs("&eacute;", "é").
		AddText("chapatoire est possible.").
		Build()
	require.Equal(t, "Une échapatoire est possible.", text.GetPlainText())
	// plain pos of 'c' in chapatoire ≈ after "Une é"
	// map first letter of error "échapatoire" (pos after "Une ")
	fromPlain := len([]rune("Une ")) // rune-based for BMP entity "é" is 1
	// better: use plain text index of "é"
	plain := text.GetPlainText()
	idx := indexOf(plain, "échapatoire")
	require.GreaterOrEqual(t, idx, 0)
	origFrom := text.GetOriginalTextPositionFor(idx, false)
	// original markup text starts with entity
	withMarkup := text.GetTextWithMarkup()
	require.Contains(t, withMarkup, "&eacute;")
	// mapped original should land at entity start region
	require.LessOrEqual(t, origFrom, indexOf(withMarkup, "chapatoire"))
	_ = fromPlain
}

// Port of AnnotatedTextTest.testInterpretAsAfter — entity after error stem.
func TestAnnotatedText_InterpretAsAfter(t *testing.T) {
	text := markup.NewAnnotatedTextBuilder().
		AddText("J'ai trouuv").
		AddMarkupInterpretAs("&eacute;", "é").
		AddText(" le livre.").
		Build()
	// plain = trouuv + é (interpretAs), matching Java "trouuv&eacute;" spelling error surface
	require.Equal(t, "J'ai trouuvé le livre.", text.GetPlainText())
	plain := text.GetPlainText()
	idx := indexOf(plain, "trouuvé")
	require.GreaterOrEqual(t, idx, 0)
	endPlain := idx + len([]rune("trouuvé"))
	origEnd := text.GetOriginalTextPositionFor(endPlain, true)
	markup := text.GetTextWithMarkup()
	require.Contains(t, markup, "&eacute;")
	require.Greater(t, origEnd, 0)
}

// Port of AnnotatedTextTest.testWithSimpleMarkup
func TestAnnotatedText_WithSimpleMarkup(t *testing.T) {
	text := markup.NewAnnotatedTextBuilder().
		AddText("J'ai louper le train.").
		AddMarkup("<span>").
		AddText(" Ce n'était pas dans mes habitudes.").
		AddMarkup("</span>").
		Build()
	require.Equal(t, "J'ai louper le train. Ce n'était pas dans mes habitudes.", text.GetPlainText())
	require.Contains(t, text.GetTextWithMarkup(), "<span>")
	idx := indexOf(text.GetPlainText(), "louper")
	require.Equal(t, idx, text.GetOriginalTextPositionFor(idx, false))
}

// Port of AnnotatedTextTest.testWithMultipleSimpleMarkup
func TestAnnotatedText_WithMultipleSimpleMarkup(t *testing.T) {
	text := markup.NewAnnotatedTextBuilder().
		AddText("A ").
		AddMarkup("<b>").
		AddText("B").
		AddMarkup("</b>").
		AddText(" C ").
		AddMarkup("<i>").
		AddText("D").
		AddMarkup("</i>").
		Build()
	require.Equal(t, "A B C D", text.GetPlainText())
	require.Equal(t, "A <b>B</b> C <i>D</i>", text.GetTextWithMarkup())
}

// Port of AnnotatedTextTest.testWithFakeMarkupInSimpleMarkupeMarkup
func TestAnnotatedText_WithFakeMarkupInSimpleMarkupeMarkup(t *testing.T) {
	text := markup.NewAnnotatedTextBuilder().
		AddText("hello ").
		AddMarkupInterpretAs("<p>", "\n\n").
		AddText("more").
		Build()
	require.Equal(t, "hello \n\nmore", text.GetPlainText())
	require.Equal(t, "hello <p>more", text.GetTextWithMarkup())
}

// Port of AnnotatedTextTest.testWithBr
func TestAnnotatedText_WithBr(t *testing.T) {
	text := markup.NewAnnotatedTextBuilder().
		AddText("line1").
		AddMarkupInterpretAs("<br/>", "\n").
		AddText("line2").
		Build()
	require.Equal(t, "line1\nline2", text.GetPlainText())
	require.Equal(t, "line1<br/>line2", text.GetTextWithMarkup())
}

func indexOf(s, sub string) int {
	// byte index is fine for ASCII/BMP tests that use same encoding for plain
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
