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
		{"ru", "Я иду в в магазин.", "RU_SOFT_V_V"},
		{"sv", "Dom är här.", "SV_SOFT_DE_DOM"},
		{"da", "Det er af af interesse.", "DA_SOFT_AF_AF"},
		{"pl", "Idę w w domu.", "PL_SOFT_W_W"},
		{"uk", "Іду в в магазин.", "UK_SOFT_V_V"},
		{"ca", "Vaig a a casa.", "CA_SOFT_A_A"},
		{"gl", "Vou de de casa.", "GL_SOFT_DE_DE"},
		{"sk", "Idem a a domov.", "SK_SOFT_A_A"},
		{"ro", "Eu de de acasă.", "RO_SOFT_DE_DE"},
		{"el", "Εγώ και και.", "EL_SOFT_KAI_KAI"},
		// remaining soft packs (token-repeat style)
		{"ar", "في في", "AR_SOFT_FI_FI"},
		{"be", "і і", "BE_SOFT_I_I"},
		{"br", "ha ha", "BR_SOFT_HA_HA"},
		{"eo", "kaj kaj", "EO_SOFT_KAJ_KAJ"},
		{"fa", "و و", "FA_SOFT_VA_VA"},
		{"ga", "agus agus", "GA_SOFT_AGUS_AGUS"},
		{"is", "og og", "IS_SOFT_OG_OG"},
		{"ja", "to to", "JA_SOFT_TO_TO"},
		{"km", "and and", "KM_SOFT_AND_AND"},
		{"lt", "ir ir", "LT_SOFT_IR_IR"},
		{"ml", "um um", "ML_SOFT_UM_UM"},
		{"sl", "in in", "SL_SOFT_IN_IN"},
		{"sr", "i i", "SR_SOFT_I_I"},
		{"ta", "um um", "TA_SOFT_UM_UM"},
		{"tl", "at at", "TL_SOFT_AT_AT"},
		{"zh", "的 的", "ZH_SOFT_DE_DE"},
		{"ast", "y y", "AST_SOFT_Y_Y"},
		{"crh", "ve ve", "CRH_SOFT_VE_VE"},
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
