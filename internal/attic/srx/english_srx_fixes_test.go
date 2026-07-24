package srx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// English segment.srx rules that need RE2 adaptations:
// - (?!(on|it|...))\p{L}{1,2} negative lookahead for "p. 6"
// - significant trailing space on ellipsis beforebreak ("... well")
func TestEnglish_PageAbbrevAndEllipsis(t *testing.T) {
	doc, err := DefaultDocument()
	require.NoError(t, err)

	// Java EnglishSRXSentenceTokenizerTest: "On p. 6 there's nothing. ", "Another sentence."
	text := "On p. 6 there's nothing. Another sentence."
	require.Equal(t, []string{"On p. 6 there's nothing. ", "Another sentence."},
		doc.Split(text, "en", "_one"))

	// "Don't split... well you know. ", "Here comes more text."
	text2 := "Don't split... well you know. Here comes more text."
	require.Equal(t, []string{"Don't split... well you know. ", "Here comes more text."},
		doc.Split(text2, "en", "_one"))
}
