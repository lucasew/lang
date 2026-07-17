package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_MultiLangSoftGrammar(t *testing.T) {
	cases := []struct {
		lang, text, rule string
	}{
		{"fr", "Je vais a la maison.", "FR_SOFT_A_LA"},
		{"es", "Yo voy haya.", "ES_SOFT_HAYA_ALLA"},
		{"pt", "Vou a o mercado.", "PT_SOFT_A_O"},
		{"it", "Vado a il negozio.", "IT_SOFT_A_IL"},
		{"nl", "Hij is als of dit.", "NL_SOFT_ALS_OF"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: tc.lang})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "grammar", f.Type)
					require.Contains(t, f.URL, "lang="+tc.lang)
					require.Contains(t, f.URL, tc.rule)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}
