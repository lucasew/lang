package languagetool

// Twin of ShortDescriptionProviderTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShortDescriptionProvider_DescriptionLength(t *testing.T) {
	p := NewShortDescriptionProvider()
	p.LoadLines = func(path string) ([]string, error) {
		return []string{"word\tshort def under forty characters"}, nil
	}
	d := p.GetShortDescription("word", "en")
	require.NotEmpty(t, d)
	require.LessOrEqual(t, len([]rune(d)), 80)
}
