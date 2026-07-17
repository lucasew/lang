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
	require.Contains(t, out, "EMPTY_LINE\tSTYLE\tstyle\t")
	require.Contains(t, out, "community.languagetool.org/rule/show/EMPTY_LINE?lang=en")
	require.Contains(t, out, "\tcore\n")
	require.Contains(t, out, "\tsoft\n")
	require.Contains(t, out, "EN_SOFT_")
	require.Contains(t, out, "# total=")
	require.Contains(t, out, "soft=")
}

func TestCoreListRules_EnUSHasSoftUSSpelling(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreListRules(&buf, "en-US"))
	out := buf.String()
	require.Contains(t, out, "EN_SOFT_COLOUR_US")
	require.Contains(t, out, "\tsoft\n")
}

func TestCoreListRules_PtBRHasRegionalSoft(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreListRules(&buf, "pt-BR"))
	out := buf.String()
	require.Contains(t, out, "PT_SOFT_AUTOCARRO_BR")
	require.Contains(t, out, "PT_SOFT_A_O") // shared pt-soft.xml still loads
	require.Contains(t, out, "\tsoft\n")
}

func TestCoreListRules_PtPTHasRegionalSoft(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreListRules(&buf, "pt-PT"))
	out := buf.String()
	require.Contains(t, out, "PT_SOFT_ONIBUS_PT")
	require.NotContains(t, out, "PT_SOFT_AUTOCARRO_BR")
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
