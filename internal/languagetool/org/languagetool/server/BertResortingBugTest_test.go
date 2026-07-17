package server

// Twin of BertResortingBugTest (Java @Ignore interactive HTTP).
// Soft: annotated markup offsets must stay in range of reconstructed source.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/stretchr/testify/require"
)

// Port of BertResortingBugTest (no active CI @Test) — position mapping smoke for #2969 class of bugs.
func TestBertResortingBug_NoTests(t *testing.T) {
	// data: three text segments with newline markup interpreted as paragraphs
	b := markup.NewAnnotatedTextBuilder().
		AddText("A teext.").
		AddMarkupInterpretAs("\n", "\n\n").
		AddText("An errör-free text.").
		AddMarkupInterpretAs("\n", "\n\n").
		AddText("So much teext.")
	at := b.Build()
	// original display string (Java: s must exactly match 'data' plain interpretation)
	s := "A teext.\nAn errör-free text.\nSo much teext."
	plain := at.GetPlainText()
	require.Contains(t, plain, "teext")
	require.Contains(t, plain, "errör-free")

	// synthetic match offsets into plain text (misspellings)
	type span struct{ from, length int }
	spans := []span{}
	// find "teext" occurrences in plain
	for i := 0; i+5 <= len([]rune(plain)); {
		r := []rune(plain)
		if string(r[i:i+5]) == "teext" {
			spans = append(spans, span{from: i, length: 5})
			i += 5
			continue
		}
		i++
	}
	require.GreaterOrEqual(t, len(spans), 2)

	// map plain offsets → original; substring must not panic
	srcRunes := []rune(s)
	for _, sp := range spans {
		origFrom := at.GetOriginalTextPositionFor(sp.from, false)
		origTo := at.GetOriginalTextPositionFor(sp.from+sp.length, true)
		require.GreaterOrEqual(t, origFrom, 0)
		require.LessOrEqual(t, origTo, len(srcRunes))
		require.LessOrEqual(t, origFrom, origTo)
		_ = string(srcRunes[origFrom:origTo]) // no panic
	}
}
