package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFuncTagger_Defaults(t *testing.T) {
	tg := FuncTagger{}
	nt := tg.CreateNullToken(" ", 0)
	require.Equal(t, " ", nt.GetToken())
	tok := tg.CreateToken("x", "VB")
	require.Equal(t, "x", tok.GetToken())
	require.Equal(t, "VB", *tok.GetPOSTag())
}
