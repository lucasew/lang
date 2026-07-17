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

func TestGolden_MultiLangSoftGrammarExtra(t *testing.T) {
	cases := []struct {
		lang, text, rule, sug string
	}{
		{"fr", "Il va a le marché.", "FR_SOFT_A_LE", "au"},
		{"fr", "Il et grand.", "FR_SOFT_ET_EST", "il est"},
		{"fr", "ca va bien.", "FR_SOFT_CA_SA", "ça va"},
		{"es", "Voy a el parque.", "ES_SOFT_A_EL", "al"},
		{"es", "Viene de el norte.", "ES_SOFT_DE_EL", "del"},
		{"es", "Ay que hacerlo.", "ES_SOFT_HAY_AY", "hay que"},
		{"pt", "Estou em o carro.", "PT_SOFT_EM_O", "no"},
		{"pt", "Livro de o autor.", "PT_SOFT_DE_O", "do"},
		{"pt", "Passou por o parque.", "PT_SOFT_POR_O", "pelo"},
		{"it", "Vado di il mercato.", "IT_SOFT_DI_IL", "del"},
		{"it", "Sono in il giardino.", "IT_SOFT_IN_IL", "nel"},
		{"nl", "Hij is te te snel.", "NL_SOFT_TE_TE", ""},
		{"nl", "Meer als gisteren.", "NL_SOFT_MEERDERE_ALS", "meer dan"},
		// second-wave packs
		{"sv", "Dom var här igår.", "SV_SOFT_DOM_VAR", "de var"},
		{"sv", "Te och och kaffe.", "SV_SOFT_OCH_OCH", ""},
		{"da", "Han kom og og gik.", "DA_SOFT_OG_OG", ""},
		{"da", "Fordi at det regner.", "DA_SOFT_FORDI_AT", ""},
		{"pl", "Kot i i pies.", "PL_SOFT_I_I", ""},
		{"pl", "Idę na na dwór.", "PL_SOFT_NA_NA", ""},
		{"ru", "Кот и и собака.", "RU_SOFT_I_I", ""},
		{"ru", "Я на на работе.", "RU_SOFT_NA_NA", ""},
		{"uk", "Кіт і і пес.", "UK_SOFT_I_I", ""},
		{"uk", "Я на на роботі.", "UK_SOFT_NA_NA", ""},
		{"ca", "Vaig de de casa.", "CA_SOFT_DE_DE", ""},
		{"ca", "Vaig a el parc.", "CA_SOFT_A_EL", "al"},
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
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}
