package el

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGreekWordTokenizerImpl(t *testing.T) {
	tok := NewGreekWordTokenizerImpl()
	got := tok.YylexTokenize("Γεια σου")
	require.NotEmpty(t, got)
}
