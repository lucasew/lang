package language

// Twin of EsperantoTest — full demo-text rules deferred; metadata smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEsperanto_LanguageSurface(t *testing.T) {
	require.Equal(t, "eo", Esperanto.GetShortCode())
	require.Equal(t, "Esperanto", Esperanto.GetName())
}
