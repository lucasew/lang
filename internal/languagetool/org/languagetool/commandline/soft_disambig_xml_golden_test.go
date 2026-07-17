package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_SoftXMLDisambig_IgnoreProductNames(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft.xml not found")
	}
	// iPhone / GitHub should not spell-flag when soft XML ignore_spelling is loaded
	for _, text := range []string{
		"I use an iPhone every day.",
		"We push code to GitHub daily.",
		"OpenAI models are popular.",
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
