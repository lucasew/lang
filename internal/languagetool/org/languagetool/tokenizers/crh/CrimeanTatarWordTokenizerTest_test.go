package crh

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrimeanTatarWordTokenizer_Tokenize(t *testing.T) {
	w := NewCrimeanTatarWordTokenizer()
	tokens := w.Tokenize("Qırımtatar Milliy Meclisiniñ 120-cı toplaşuvı olıp keçti")
	require.Equal(t, "[Qırımtatar,  , Milliy,  , Meclisiniñ,  , 120-cı,  , toplaşuvı,  , olıp,  , keçti]",
		"["+strings.Join(tokens, ", ")+"]")
}
