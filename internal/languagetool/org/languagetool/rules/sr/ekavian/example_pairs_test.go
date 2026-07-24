package ekavian

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSR_Ekavian_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"битке"}, NewMorfologikEkavianSpellerRule().GetIncorrectExamples()[0].GetCorrections())
}
