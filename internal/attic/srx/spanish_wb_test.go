package srx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Spanish segment.srx: break=no on \b(…|...)[\p{Pe}»…][\s] before lowercase.
// RE2 cannot express Java Unicode \b; consuming expansion missed letter→ellipsis
// (e.g. "tal…» sale"). Empty () + runtime isJavaWordBoundary is required.
func TestSpanish_EllipsisGuillemet_NoSplit(t *testing.T) {
	doc, err := DefaultDocument()
	require.NoError(t, err)
	text := "«El tal del tal…» sale de aquí"
	require.Equal(t, []string{text}, doc.Split(text, "es", "_two"))
}
