package language

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Portuguese rule priorities from org.languagetool.language.Portuguese (id2prio + getPriorityForId).
// Java is king — do not invent extra IDs.

var portugueseID2Prio = map[string]int{
	"ARCHAISMS": -26,
	"AUSENCIA_VIRGULA": 1,
	"AUX_VERBO": -45,
	"BARBARISMS_PT_PT_V4": -10,
	"BIASED_OPINION_WORDS": -31,
	"CHILDISH_LANGUAGE": -25,
	"COLOCAÇÃO_ADVÉRBIO": -90,
	"CONFUSION_POR_PÔR_V2": 10,
	"CONTA_TO": -44,
	"CRASE_CONFUSION": -54,
	"DEGREE_MINUTES_SECONDS": 30,
	"DIACRITICS": -45,
	"EMAIL": 1,
	"EMAIL_SEM_HIFEN": -45,
	"ENSINO_A_DISTANCIA": -45,
	"EU_NÓS_REMOVAL": -90,
	"FAZER_USO_DE-USAR-RECORRER": -90,
	"FILLER_WORDS_PT": -990,
	"FINAL_STOPS": -75,
	"FORMAL_T-V_DISTINCTION": -100,
	"FORMAL_T-V_DISTINCTION_ALL": -101,
	"FRAGMENT_TWO_ARTICLES": 50,
	"GENERAL_GENDER_NUMBER_AGREEMENT_ERRORS": -56,
	"GENERAL_NUMBER_AGREEMENT_ERRORS": -56,
	"GENERAL_VERB_AGREEMENT_ERRORS": -55,
	"HOMOPHONE_AS_CARD": 5,
	"INFORMALITIES": -27,
	"INTERJECTIONS_PUNTUATION": 20,
	"INTERNET_ABBREVIATIONS": -24,
	"LP_PARONYMS": 10,
	"NAO_MILITARES_CIVIS": -54,
	"NA_NÃO": 10,
	"NA_QUELE": -54,
	"NOTAS_FICAIS": -54,
	"OQ_O_QUE_ORTHOGRAPHY": -45,
	"PARONYM_CRITICA_397": 10,
	"PARONYM_INICIO_169": 10,
	"PARONYM_MUSICO_499_bis": 10,
	"PARONYM_POLITICA_523": 10,
	"PARONYM_PRONUNCIA_262": 10,
	"PRETERITO_PERFEITO": -51,
	"PROFANITY": -6,
	"PT_AGREEMENT_REPLACE": -35,
	"PT_BARBARISMS_REPLACE": -10,
	"PT_BR_SIMPLE_REPLACE": -51,
	"PT_CLICHE_REPLACE": -17,
	"PT_COMPOUNDS_POST_REFORM": -45,
	"PT_DIACRITICS_REPLACE": -45,
	"PT_ENGLISH_CONTRACTION_ORTHOGRAPHY": -45,
	"PT_PT_SIMPLE_REPLACE": -11,
	"PT_REDUNDANCY_REPLACE": -12,
	"PT_WIKIPEDIA_COMMON_ERRORS": -500,
	"PT_WORDINESS_REPLACE": -13,
	"READABILITY_RULE_DIFFICULT_PT": -1101,
	"READABILITY_RULE_SIMPLE_PT": -1100,
	"REPEATED_WORDS": -210,
	"TODOS_FOLLOWED_BY_NOUN_PLURAL": 3,
	"TODOS_FOLLOWED_BY_NOUN_SINGULAR": 2,
	"UNKNOWN_WORD": -2000,
	"UNPAIRED_BRACKETS": -5,
	"UPPERCASE_SENTENCE_START": -600,
	"VERB_COMMA_CONJUNCTION": 10,
}

// PortuguesePriorityMap ports Portuguese.getPriorityMap (defensive copy).
func PortuguesePriorityMap() map[string]int {
	out := make(map[string]int, len(portugueseID2Prio))
	for k, v := range portugueseID2Prio {
		out[k] = v
	}
	return out
}

// PortuguesePriorityForId ports Portuguese.getPriorityForId (then Language base).
// Java applies several prefix checks BEFORE the id2prio map.
func PortuguesePriorityForId(id string) int {
	if strings.HasPrefix(id, "MORFOLOGIK_RULE") {
		return -50
	}
	if strings.HasPrefix(id, "PT_SIMPLE_REPLACE_ORTHOGRAPHY") {
		return -49
	}
	if strings.HasPrefix(id, "AI_PT_GGEC_REPLACEMENT_ORTHOGRAPHY_SPELL") {
		return -48
	}
	if strings.HasPrefix(id, "PT_MULTITOKEN_SPELLING") {
		return -48
	}
	if strings.HasPrefix(id, "AI_PT_GGEC_REPLACEMENT_OTHER") {
		return -4
	}
	if strings.HasPrefix(id, "ACENTUAÇÃO_VOGAL_ÊNCLISE") {
		return -51
	}
	if strings.HasPrefix(id, "COLOCACAO_PRONOMINAL_COM_ATRATOR") {
		return -52
	}
	if p, ok := portugueseID2Prio[id]; ok {
		return p
	}
	if strings.HasPrefix(id, "AI_PT_HYDRA_LEO") {
		// Java MISSING_COMMA branch and residual both return -51
		return -51
	}
	return languagePriorityForId(id)
}

// PortuguesePrepareLineForSpeller ports Portuguese.prepareLineForSpeller.
func PortuguesePrepareLineForSpeller(line string) []string {
	parts := strings.Split(line, "#")
	if len(parts) == 0 {
		return []string{line}
	}
	formTag := regexp.MustCompile(`[\t;]`).Split(parts[0], -1)
	// Java: formTag[i].trim()
	form := tools.JavaStringTrim(formTag[0])
	if len(formTag) > 1 {
		tag := tools.JavaStringTrim(formTag[1])
		if strings.HasPrefix(tag, "N") || tag == "_Latin_" {
			return []string{form}
		}
		return []string{""}
	}
	return []string{line}
}
