package commandline

import (
	"bytes"
	"encoding/json"
	"strings"
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
	// soft listed before core; footer soft breakdown
	softIdx := strings.Index(out, "\tsoft\n")
	coreIdx := strings.Index(out, "\tcore\n")
	require.True(t, softIdx >= 0 && coreIdx >= 0 && softIdx < coreIdx, "soft should precede core")
	require.Contains(t, out, "soft_grammar=")
	require.Contains(t, out, "soft_style=")
	require.Contains(t, out, "soft_typographical=")
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

func TestCoreListRules_EsMXHasRegionalSoft(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreListRules(&buf, "es-MX"))
	out := buf.String()
	require.Contains(t, out, "ES_SOFT_ORDENADOR_MX")
	require.NotContains(t, out, "ES_SOFT_COMPUTADORA_ES")
}

func TestCoreListRules_DeCHHasRegionalSoft(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreListRules(&buf, "de-CH"))
	out := buf.String()
	require.Contains(t, out, "DE_SOFT_STRASSE_CH")
	require.Contains(t, out, "DE_SOFT_DAS_DASS") // shared de-soft.xml
}

func TestCoreListRules_FrCAHasRegionalSoft(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreListRules(&buf, "fr-CA"))
	out := buf.String()
	require.Contains(t, out, "FR_SOFT_WEEKEND_CA")
	require.Contains(t, out, "FR_SOFT_A_LA")
}

func TestCoreListRules_PickySoftOnlyWhenPicky(t *testing.T) {
	var def bytes.Buffer
	require.NoError(t, CoreListRules(&def, "en"))
	require.NotContains(t, def.String(), "EN_SOFT_PICKY_UTILIZE")

	// list-rules uses configureCoreLT without picky; verify picky path via golden hook registration
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Please utilize the tool.", &CommandLineOptions{
		Language: "en",
		Level:    "PICKY",
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_PICKY_UTILIZE" {
			found = true
			require.Equal(t, "style", f.Type)
			require.Equal(t, "use", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
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
