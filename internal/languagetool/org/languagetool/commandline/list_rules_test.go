package commandline

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreListRules(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreListRules(&buf, "en"))
	out := buf.String()
	require.Contains(t, out, "EN_A_VS_AN")
	require.Contains(t, out, "GRAMMAR")
}

func TestRunWithIO_RulesSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"rules", "-l", "en"}, DefaultCoreHooks(), &out, &errb)
	require.Equal(t, 0, code, errb.String())
	require.Contains(t, out.String(), "EN_A_VS_AN")
}

func TestNormalizeProductArgs_Rules(t *testing.T) {
	require.Equal(t, []string{"--list-rules", "-l", "de"}, NormalizeProductArgs([]string{"rules", "-l", "de"}))
}
