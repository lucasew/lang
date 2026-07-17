package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_SoftXMLDisambig_IgnoreProductNames(t *testing.T) {
	if DiscoverEnglishIgnoreSpellingList(nil) == "" && DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("soft disambiguation data not found")
	}
	// product/tech names from en-ignore-spelling.txt should not spell-flag
	for _, text := range []string{
		"I use an iPhone every day.",
		"We push code to GitHub daily.",
		"OpenAI models are popular.",
		"We deploy with Kubernetes and Docker.",
		"I prefer TypeScript over JavaScript.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftXML_WillRunFilter(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" || DiscoverEnglishPOSDict(nil) == "" {
		t.Skip("need en-soft.xml and english.dict")
	}
	var out bytes.Buffer
	err := CoreTagHook(&out, "I will run tomorrow.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	s := out.String()
	// run should show VB (possibly only VB after filter)
	require.Contains(t, s, "run/")
	require.Contains(t, s, "VB")
}
