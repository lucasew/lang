package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOptions_IgnoreWords(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{"-l", "en", "--ignore-words", "xyzzy,Foobar", "-"})
	require.NoError(t, err)
	require.Equal(t, []string{"xyzzy", "Foobar"}, opts.GetIgnoreWords())
}

func TestGolden_IgnoreWordsSuppressesSpelling(t *testing.T) {
	// without ignore: "xyzzy" should spell-flag under binary or demo speller
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I saw xyzzy today.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	hadSpell := false
	for _, f := range findings {
		if f.Rule == "MORFOLOGIK_RULE_EN_US" {
			hadSpell = true
		}
	}
	if !hadSpell {
		t.Skip("no binary/demo speller active for xyzzy")
	}

	buf.Reset()
	_, err = CoreGoldenHook(&buf, "I saw xyzzy today.", &CommandLineOptions{
		Language:    "en",
		IgnoreWords: []string{"xyzzy"},
	})
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestRunWithIO_IgnoreWords(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--ignore-words", "xyzzy", "--lint", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I saw xyzzy today.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1, "code=%d err=%s", code, errb.String())
	require.NotContains(t, out.String(), "MORFOLOGIK_RULE_EN_US")
}
