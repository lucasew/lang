package srx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Catalan composed abbrevs Dra. Ma. / Sta. Ma. need loomchild exception
// lookbehind semantics (forward Find prefers short alt Dra over Dra. Ma).
func TestCatalan_DraMa_StaMa(t *testing.T) {
	doc, err := DefaultDocument()
	require.NoError(t, err)
	require.Equal(t, []string{"La Dra. Ma. Victòria."},
		doc.Split("La Dra. Ma. Victòria.", "ca", "_one"))
	require.Equal(t, []string{"la projectada Sta. Ma. de Gàllecs"},
		doc.Split("la projectada Sta. Ma. de Gàllecs", "ca", "_one"))
	require.Equal(t, []string{"La Dra. Victòria."},
		doc.Split("La Dra. Victòria.", "ca", "_one"))
}
