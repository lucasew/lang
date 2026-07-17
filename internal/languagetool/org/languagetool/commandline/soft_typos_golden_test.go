package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscoverEnglishTyposFile(t *testing.T) {
	p := DiscoverEnglishTyposFile(nil)
	if p == "" {
		t.Skip("en-typos.tsv not found")
	}
	require.FileExists(t, p)
}

func TestGolden_SoftTyposSuggestions(t *testing.T) {
	if DiscoverEnglishTyposFile(nil) == "" && DiscoverEnglishUSDict(nil) == "" {
		t.Skip("need typos file or binary dict")
	}
	cases := []struct{ text, sug string }{
		{"I will go tommorow.", "tomorrow"},
		{"That is wierd.", "weird"},
		{"Please recieve this.", "receive"},
	}
	for _, tc := range cases {
		t.Run(tc.sug, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == "MORFOLOGIK_RULE_EN_US" && f.Suggestion == tc.sug {
					found = true
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftCanCan(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "They can can fish.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_CAN_CAN" {
			found = true
			require.Equal(t, "can", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftThatThat(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I know that that is true.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_THAT_THAT" {
			found = true
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftHadOf(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I had of known better.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_HAD_OF" {
			found = true
			require.Equal(t, "had", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}
