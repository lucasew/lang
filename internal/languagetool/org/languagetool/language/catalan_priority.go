package language

import (
	"regexp"
	"strings"
)

// Catalan rule priorities from org.languagetool.language.Catalan.getPriorityForId.
// Switch-table port as map (exact id equality). Java is king — do not invent IDs.

var catalanPriorityExact = map[string]int{
	"ACCENTUACIO": 5,
	"ACCENTUATION_CHECK": 10,
	"AGREEMENT_POSTPONED_ADJ": -15,
	"AMB_EM": 5,
	"APOSTROFACIO_MOT_DESCONEGUT": -120,
	"APOSTROF_ANYS": 5,
	"ARRIBAN_ARRIBANT": 30,
	"ARTICLE_TOPONIM_MIN": -10,
	"AVIS": 10,
	"A_PER": 10,
	"CAP_ELS_CAP_ALS": 10,
	"CAP_GENS": 20,
	"CASING": 10,
	"CASING_START": -5,
	"CA_END_PARAGRAPH_PUNCTUATION": -250,
	"CA_REMOTE_RULE": 15,
	"CA_SPLIT_LONG_SENTENCE": -90,
	"CA_SPLIT_WORDS": 30,
	"CA_WORD_COHERENCY": -10,
	"CA_WORD_COHERENCY_VALENCIA": -10,
	"COMETES_INCORRECTES": 50,
	"COMMA_ENTRE_DALTRES": 20,
	"COMMA_IJ": 10,
	"COMMA_LOCUTION": -10,
	"COMMA_PERO1": 35,
	"CONCORDANCA_GRIS": 10,
	"CONCORDANCA_PRONOMS_CATCHALL": -10,
	"CONCORDANCES_CASOS_PARTICULARS": 30,
	"CONCORDANCES_DET_ADJ": 5,
	"CONCORDANCES_DET_NOM": 5,
	"CONCORDANCES_DET_POSSESSIU": 5,
	"CONCORDANCES_NOUNS_PRIORITY": 10,
	"CONCORDANCES_NUMERALS": 10,
	"CONCORDANCES_NUMERALS_DUES": 10,
	"CONEIXENTS": 40,
	"CONEIXET": 40,
	"CONEIXO_CONEC": 50,
	"CONFUSIONS": 30,
	"CONFUSIONS2": 80,
	"CONFUSIONS_ACCENT": 20,
	"CONFUSIONS_PRONOMS_FEBLES": 35,
	"CONFUSIO_PASSAT_INFINITIU": 20,
	"CONTRACCIONS": 0,
	"DESDE_UN": 40,
	"DET_GN": 5,
	"DEUS_SEUS": 5,
	"DEU_NI_DO": 80,
	"DIACRITICS": 20,
	"DICENDI_QUE": -250,
	"DOS_ARTICLES": 10,
	"ELA_GEMINADA": 35,
	"ELA_GEMINADA_WIKI": -300,
	"EL_FAN_AGENOLLAR": 10,
	"EN_NO_INFINITIU_CAUSAL_REMOTE": 5,
	"ESPAIS_QUE_FALTEN_PUNTUACIO": -20,
	"ESPAIS_SOBRANTS": 40,
	"ESPERANT_US_AGRADI": 40,
	"ES_UNKNOWN": 25,
	"ET_AL": 30,
	"EXIGEIX_ACCENTUACIO_VALENCIANA": -120,
	"FALTA_COMA_FRASE_CONDICIONAL": -20,
	"FALTA_CONDICIONAL": 10,
	"FALTA_ELEMENT_ENTRE_VERBS": -200,
	"FER_LOGIN": 70,
	"FIDEUA": 5,
	"GERUNDI_PERD_T": 30,
	"HAVER_PARTICIPI_HAVER_IMPERSONAL": 15,
	"HAVER_SENSE_HAC": 25,
	"HA_A": 25,
	"INCORRECT_EXPRESSIONS": 50,
	"INCORRECT_WORDS_IN_CONTEXT": 28,
	"LO_NEUTRE": 40,
	"L_D_N_NO_S_APOSTROFEN": 5,
	"L_NO_APOSTROFA": 5,
	"L_OK": 70,
	"MAJOR_MES_GRAN0": -40,
	"MAJUSCULA_IMPROBABLE": -300,
	"MORFOLOGIK_RULE_CA_ES": -100,
	"MOTS_GUIONET": 10,
	"MOTS_NO_SEPARATS": 40,
	"MOTS_SENSE_GUIONETS": 20,
	"MUNDAR": -50,
	"NOMBRES_ROMANS": -90,
	"OBLIDARSE": 30,
	"OFERTAR_OFERIR": 50,
	"ORDINALS": 20,
	"PASSAR_SE": 35,
	"PASSAT_PERIFRASTIC": 25,
	"PEL_QUE": -10,
	"PERO_PERO": 30,
	"PERSONATGES_FAMOSOS": 50,
	"PER_A_QUE_PERQUE": 40,
	"PHRASE_REPETITION": -150,
	"PORTA_UNA_HORA": -40,
	"PORTO_LLEGINT": -30,
	"POSTULARSE": 10,
	"PREFIXOS_SENSE_GUIONET_EN_DICCIONARI": 10,
	"PREGUEM_DISCULPIN": 45,
	"PREPOSICIONS_MINUSCULA": -97,
	"PREPOSITIONS": 25,
	"PRONOMS_FEBLES_COLLOQUIALS": 30,
	"PRONOMS_FEBLES_COMBINACIONS_SE": 40,
	"PRONOMS_FEBLES_DARRERE_VERB": 30,
	"PRONOMS_FEBLES_SOLTS": -10,
	"PRONOMS_FEBLES_SOLTS2": 26,
	"PRONOMS_FEBLES_TEMPS_VERBAL": 35,
	"PRONOM_FEBLE_HI": 20,
	"PUNCTUATION_PARAGRAPH_END": -200,
	"PUNT_FINAL": -200,
	"PUNT_LLETRA": 30,
	"QUAN_PREPOSICIO": -10,
	"RECENT": 10,
	"REEMPRENDRE": 28,
	"REGIONAL_VERBS": -10,
	"REPETEAD_ELEMENTS": 40,
	"REPETITIONS_STYLE": -50,
	"REPETITION_ADJ_N_ADJ": -155,
	"SELS_EN_VA": 10,
	"SELS_EN_VA_DE_LES_MANS": 10,
	"SE_LI_VA_FER_CALLAR": 15,
	"SON_BONIC": 5,
	"SPELLING": 5,
	"SUBSTANTIUS_JUNTS": -150,
	"SUGGERIMENTS_LE": -97,
	"SUPER": 20,
	"TASCAS_TASQUES": -97,
	"TENIR_QUE": 35,
	"UN_ALTRE_DISTRIBUTIVES": -10,
	"UPPERCASE_SENTENCE_START": -300,
	"URL": 10,
	"VENIR_NO_REFLEXIU": 5,
	"VERBS_NOMSPROPIS": -20,
	"VERBS_NO_INCOATIUS": 30,
	"VERBS_PRONOMINALS": -25,
	"ZERO_O": 10,
}

// catalanSpellerExceptions ports Catalan.spellerExceptions.
var catalanSpellerExceptions = map[string]struct{}{
	"San Juan": {},
	"Copa América": {},
	"Colección Jumex": {},
	"Banco Santander": {},
	"San Marcos": {},
	"Santa Ana": {},
	"San Joaquín": {},
	"Naguib Mahfouz": {},
	"Rosalía": {},
	"Aristide Maillol": {},
	"Alexia Putellas": {},
	"Mónica Randall": {},
	"Vicente Blasco Ibáñez": {},
	"Copa Sudamericana": {},
	"Série A": {},
	"Banco Sabadell": {},
}

// CatalanPriorityExactMap returns a defensive copy of the exact-id priority table.
func CatalanPriorityExactMap() map[string]int {
	out := make(map[string]int, len(catalanPriorityExact))
	for k, v := range catalanPriorityExact {
		out[k] = v
	}
	return out
}

// CatalanPriorityForId ports Catalan.getPriorityForId (then Language base).
// Prefix checks run after exact switch (Java order).
func CatalanPriorityForId(id string) int {
	if p, ok := catalanPriorityExact[id]; ok {
		return p
	}
	if strings.HasPrefix(id, "CA_MULTITOKEN_SPELLING") {
		return -95
	}
	if strings.HasPrefix(id, "CA_SIMPLE_REPLACE_MULTIWORDS") {
		return 70
	}
	if strings.HasPrefix(id, "CA_SIMPLE_REPLACE_ANGLICISM") {
		return 65 // greater than CA_SIMPLE_REPLACE_BALEARIC
	}
	if strings.HasPrefix(id, "CA_SIMPLE_REPLACE_BALEARIC") {
		return 60
	}
	if strings.HasPrefix(id, "CA_SIMPLE_REPLACE_VERBS") {
		return 28
	}
	if strings.HasPrefix(id, "CA_COMPOUNDS") {
		return 50
	}
	if strings.HasPrefix(id, "CA_SIMPLE_REPLACE_DIACRITICS_IEC") {
		return 0
	}
	if strings.HasPrefix(id, "CA_SIMPLE_REPLACE") {
		return 30
	}
	return languagePriorityForId(id)
}

// CatalanPrepareLineForSpeller ports Catalan.prepareLineForSpeller.
func CatalanPrepareLineForSpeller(line string) []string {
	parts := strings.Split(line, "#")
	if len(parts) == 0 {
		return []string{line}
	}
	formTag := regexp.MustCompile(`[\t;]`).Split(parts[0], -1)
	form := strings.TrimSpace(formTag[0])
	if _, bad := catalanSpellerExceptions[form]; bad {
		return []string{""}
	}
	if len(formTag) > 1 {
		tag := strings.TrimSpace(formTag[1])
		if strings.HasPrefix(tag, "N") || tag == "_Latin_" {
			return []string{form}
		}
		return []string{""}
	}
	return []string{line}
}
