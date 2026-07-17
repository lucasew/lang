package commandline

import (
	"bytes"
	"encoding/json"
	"strings"
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
		{"fr", "Ils son partis.", "FR_SOFT_SON_SONT", "ils sont"},
		{"fr", "Elle et grande.", "FR_SOFT_ELLE_ET", "elle est"},
		{"fr", "Je reste parceque il pleut.", "FR_SOFT_PARCE_QUE", "parce que"},
		{"fr", "Bravo!!", "FR_SOFT_DOUBLE_BANG", ""},
		{"es", "Voy a el parque.", "ES_SOFT_A_EL", "al"},
		{"es", "Viene de el norte.", "ES_SOFT_DE_EL", "del"},
		{"es", "Ay que hacerlo.", "ES_SOFT_HAY_AY", "hay que"},
		{"es", "Él a ido ya.", "ES_SOFT_HA_A", "ha ido"},
		{"es", "Lo hago por que quiero.", "ES_SOFT_PORQUE", ""},
		{"es", "Hola!!", "ES_SOFT_DOUBLE_BANG", ""},
		{"pt", "Estou em o carro.", "PT_SOFT_EM_O", "no"},
		{"pt", "Livro de o autor.", "PT_SOFT_DE_O", "do"},
		{"pt", "Passou por o parque.", "PT_SOFT_POR_O", "pelo"},
		{"it", "Vado di il mercato.", "IT_SOFT_DI_IL", "del"},
		{"it", "Sono in il giardino.", "IT_SOFT_IN_IL", "nel"},
		{"it", "Vado a lo zoo.", "IT_SOFT_A_LO", "allo"},
		{"it", "Libro di lo studente.", "IT_SOFT_DI_LO", "dello"},
		{"it", "Perche sei qui?", "IT_SOFT_PERCHE", "perché"},
		{"it", "Ciao!!", "IT_SOFT_DOUBLE_BANG", "!"},
		{"pt", "Estou em a casa.", "PT_SOFT_EM_A", "na"},
		{"pt", "Livro de a autora.", "PT_SOFT_DE_A", "da"},
		{"pt", "Vou a a praia.", "PT_SOFT_A_A", ""},
		{"pt", "Olá!!", "PT_SOFT_DOUBLE_BANG", "!"},
		{"nl", "Hij is te te snel.", "NL_SOFT_TE_TE", ""},
		{"nl", "Meer als gisteren.", "NL_SOFT_MEERDERE_ALS", "meer dan"},
		{"nl", "Hij is groter als mij.", "NL_SOFT_GROTER_ALS", "groter dan"},
		{"nl", "Brood en en kaas.", "NL_SOFT_EN_EN", ""},
		{"nl", "Hun hebben gelijk.", "NL_SOFT_HUN_HEBBEN", ""},
		{"nl", "Goed!!", "NL_SOFT_DOUBLE_BANG", "!"},
		// second-wave packs
		{"sv", "Dom var här igår.", "SV_SOFT_DOM_VAR", "de var"},
		{"sv", "Te och och kaffe.", "SV_SOFT_OCH_OCH", ""},
		{"sv", "Dom har en bil.", "SV_SOFT_DOM_HAR", "de har"},
		{"sv", "Han är större som mig.", "SV_SOFT_STORRE_AN", "större än"},
		{"sv", "Medans vi väntar.", "SV_SOFT_MEDANS", "medan"},
		{"sv", "Hej!!", "SV_SOFT_DOUBLE_BANG", "!"},
		{"da", "Han kom og og gik.", "DA_SOFT_OG_OG", ""},
		{"da", "Fordi at det regner.", "DA_SOFT_FORDI_AT", ""},
		{"da", "Han er større som mig.", "DA_SOFT_STORRE_END", "større end"},
		{"da", "Dem har en bil.", "DA_SOFT_HUN_HAR", "de har"},
		{"da", "Det er ikke ikke sandt.", "DA_SOFT_IKKE_IKKE", ""},
		{"da", "Hej!!", "DA_SOFT_DOUBLE_BANG", "!"},
		{"pl", "Kot i i pies.", "PL_SOFT_I_I", ""},
		{"pl", "Idę na na dwór.", "PL_SOFT_NA_NA", ""},
		{"pl", "To nie nie działa.", "PL_SOFT_NIE_NIE", ""},
		{"pl", "Idę z z domu.", "PL_SOFT_Z_Z", ""},
		{"pl", "Cześć!!", "PL_SOFT_DOUBLE_BANG", "!"},
		{"ru", "Кот и и собака.", "RU_SOFT_I_I", ""},
		{"ru", "Я на на работе.", "RU_SOFT_NA_NA", ""},
		{"ru", "Он с с другом.", "RU_SOFT_S_S", ""},
		{"ru", "Это это важно.", "RU_SOFT_ETO_ETO", ""},
		{"ru", "Привет!!", "RU_SOFT_DOUBLE_BANG", "!"},
		{"uk", "Кіт і і пес.", "UK_SOFT_I_I", ""},
		{"uk", "Я на на роботі.", "UK_SOFT_NA_NA", ""},
		{"uk", "Він з з дому.", "UK_SOFT_Z_Z", ""},
		{"uk", "Це це добре.", "UK_SOFT_TSE_TSE", ""},
		{"uk", "Привіт!!", "UK_SOFT_DOUBLE_BANG", "!"},
		{"ca", "Vaig de de casa.", "CA_SOFT_DE_DE", ""},
		{"ca", "Vaig a el parc.", "CA_SOFT_A_EL", "al"},
		{"ca", "Ve de el nord.", "CA_SOFT_DE_EL", "del"},
		{"ca", "Anem en en cotxe.", "CA_SOFT_EN_EN", ""},
		{"ca", "Hola!!", "CA_SOFT_DOUBLE_BANG", "!"},
		{"gl", "Vou a a casa.", "GL_SOFT_A_A", ""},
		{"gl", "Vou a o mercado.", "GL_SOFT_A_O", "ao"},
		{"gl", "Libro de o autor.", "GL_SOFT_DE_O", "do"},
		{"gl", "Vou en en tren.", "GL_SOFT_EN_EN", ""},
		{"gl", "Ola!!", "GL_SOFT_DOUBLE_BANG", "!"},
		{"sk", "Pes i i mačka.", "SK_SOFT_I_I", ""},
		{"sk", "Idem na na dvor.", "SK_SOFT_NA_NA", ""},
		{"sk", "Som v v meste.", "SK_SOFT_V_V", ""},
		{"sk", "To nie nie je.", "SK_SOFT_NIE_NIE", ""},
		{"sk", "Ahoj!!", "SK_SOFT_DOUBLE_BANG", "!"},
		{"ro", "Pisică și și câine.", "RO_SOFT_SI_SI", ""},
		{"ro", "Merg în în casă.", "RO_SOFT_IN_IN", ""},
		{"ro", "Merg pe pe drum.", "RO_SOFT_PE_PE", ""},
		{"ro", "Vin cu cu el.", "RO_SOFT_CU_CU", ""},
		{"ro", "Salut!!", "RO_SOFT_DOUBLE_BANG", "!"},
		{"el", "Το το βιβλίο.", "EL_SOFT_TO_TO", ""},
		{"el", "Θέλω να να φύγω.", "EL_SOFT_NA_NA", ""},
		{"el", "Μιλάω με με φίλο.", "EL_SOFT_ME_ME", ""},
		{"el", "Δεν δεν ξέρω.", "EL_SOFT_DEN_DEN", ""},
		{"el", "Γεια!!", "EL_SOFT_DOUBLE_BANG", "!"},
		{"sl", "On je je tukaj.", "SL_SOFT_JE_JE", ""},
		{"sr", "Idem u u školu.", "SR_SOFT_U_U", ""},
		{"lt", "Katė ir ir šuo.", "LT_SOFT_IR_IR", ""},
		{"is", "Kaffi og og te.", "IS_SOFT_OG_OG", ""},
		{"eo", "La la libro.", "EO_SOFT_LA_LA", ""},
		{"br", "Ha ha gant.", "BR_SOFT_HA_HA", ""},
		{"ga", "An an madra.", "GA_SOFT_AN_AN", ""},
		{"zh", "是 是 的。", "ZH_SOFT_SHI_SHI", ""},
		{"ja", "no no desu.", "JA_SOFT_NO_NO", ""},
		{"ast", "casa de de madera.", "AST_SOFT_DE_DE", ""},
		{"be", "кот на на стале.", "BE_SOFT_NA_NA", ""},
		{"crh", "bir de de eki.", "CRH_SOFT_DE_DE", ""},
		{"km", "the the book", "KM_SOFT_THE_THE", ""},
		{"ml", "or or not", "ML_SOFT_OR_OR", ""},
		{"ta", "and and more", "TA_SOFT_AND_AND", ""},
		{"tl", "sa sa bahay", "TL_SOFT_SA_SA", ""},
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
					if strings.Contains(tc.rule, "DOUBLE_BANG") || strings.Contains(tc.rule, "DOUBLE_Q") {
						require.Equal(t, "typographical", f.Type)
					} else {
						require.Equal(t, "grammar", f.Type)
					}
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}
