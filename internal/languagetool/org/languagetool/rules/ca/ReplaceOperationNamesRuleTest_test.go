package ca

// Twin of ReplaceOperationNamesRuleTest — POS-gated Match (Java FreeLing tags injected).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	catok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ca"
	"github.com/stretchr/testify/require"
)

// opNamesTagWord injects FreeLing-style tags used by ReplaceOperationNamesRule gates.
// Only tokens that participate in prev/next/exception checks need tags.
func opNamesTagWord(token string) []languagetool.TokenTag {
	low := strings.ToLower(token)
	switch low {
	// determiners D[^R].* / D[^R].M.*
	case "el", "l'", "l":
		return []languagetool.TokenTag{{POS: "DA0MS0", Lemma: "el"}}
	case "els":
		return []languagetool.TokenTag{{POS: "DA0MP0", Lemma: "el"}}
	case "la":
		return []languagetool.TokenTag{{POS: "DA0FS0", Lemma: "el"}}
	case "les":
		return []languagetool.TokenTag{{POS: "DA0FP0", Lemma: "el"}}
	case "un":
		return []languagetool.TokenTag{{POS: "DI0MS0", Lemma: "un"}}
	case "una":
		return []languagetool.TokenTag{{POS: "DI0FS0", Lemma: "un"}}
	// prepositions SPS00
	case "d'", "de", "del", "amb", "per", "pel", "com", "des", "en", "a", "al":
		return []languagetool.TokenTag{{POS: "SPS00", Lemma: low}}
	// nouns N.* (suppress when prev/next is noun)
	case "llibre", "informe", "procés", "resultat", "equip", "batalla", "braç", "riu",
		"vi", "ministre", "rodes", "ampolles", "llibres", "cotes", "cervell", "claus",
		"matrimoni", "cafè", "cos", "característiques", "firmes", "comercials", "marca",
		"mercat", "mort", "vegades", "pressa", "llum", "nostàlgia", "duplicat":
		return []languagetool.TokenTag{{POS: "NCMS000", Lemma: low}}
	case "empaquetat", "equilibrat", "filtrat", "etiquetat", "embotellat", "tenyit",
		"assecat", "processat", "rentat", "relligat", "repicat", "rejuntat", "polit",
		"liderat", "observat":
		// participial / operation names often tagged as adjectives or participles;
		// not required for the match itself (lookup is surface). Leave untagged unless needed.
		return nil
	// adverbs / RG that block as prev exception
	case "ben", "molt", "bastant", "tot", "no", "llavors", "diverses", "moltes", "una vegada":
		return []languagetool.TokenTag{{POS: "RG", Lemma: low}}
	case "van", "és", "era", "ha", "tinc", "eixia", "arribi", "fer":
		return []languagetool.TokenTag{{POS: "VMIP3S0", Lemma: low}}
	case "i", "o":
		return []languagetool.TokenTag{{POS: "CC", Lemma: low}}
	case ".":
		return []languagetool.TokenTag{{POS: "PUNCT", Lemma: low}}
	default:
		return nil
	}
}

func analyzeOpNames(s string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTaggerAndTokenizer(s, opNamesTagWord, catok.NewCatalanWordTokenizer())
}

func TestReplaceOperationNamesRule_Rule(t *testing.T) {
	rule := NewReplaceOperationNamesRule(nil)

	// correct sentences (Java twin) — POS inject suppresses false positives
	for _, s := range []string{
		"tot tenyit amb llum de nostàlgia",
		"Ho van fer per duplicat.",
		"Assecat el braç del riu",
		"el llibre empaquetat",
		"un resultat equilibrat",
		"el nostre equip era bastant equilibrat",
		"un llibre ben empaquetat",
		"l'informe filtrat pel ministre",
		"L'informe filtrat és terrible",
		"ha liderat la batalla",
		"Els tinc empaquetats",
		"amb tractament unitari i equilibrat",
		"Processat després de la mort de Carles II",
		"Processat diverses vegades",
		"moltes vegades empaquetat amb pressa",
		"és llavors embotellat i llançat al mercat",
		"la comercialització de vi embotellat amb les firmes comercials",
		"eixia al mercat el vi blanc embotellat amb la marca",
		"que arribi a un equilibrat matrimoni",
		"És un cafè amb molt de cos i molt equilibrat.",
		"i per tant etiquetat com a observat",
		"Molt equilibrat en les seves característiques",
		"filtrat per Wikileaks",
		"una vegada filtrat",
		"no equilibrat",
	} {
		matches := rule.Match(analyzeOpNames(s))
		require.Empty(t, matches, "correct: %s", s)
	}

	// incorrect (Java twin)
	for _, s := range []string{
		"Assecat del braç del riu",
		"Cal vigilar el filtrat del vi",
		"El procés d'empaquetat",
		"El procés d'etiquetat de les ampolles",
		"El rentat de cotes",
	} {
		matches := rule.Match(analyzeOpNames(s))
		require.NotEmpty(t, matches, "incorrect: %s", s)
	}

	// plurals need synthesizer
	rule.Synthesize = func(tok *languagetool.AnalyzedToken, postagRE string) []string {
		if tok == nil || tok.GetLemma() == nil {
			return nil
		}
		lemma := *tok.GetLemma()
		// minimal plural forms for twin assertions (Java CatalanSynthesizer NC.P.*)
		switch lemma {
		case "rentada":
			return []string{"rentades"}
		case "rentatge":
			return []string{"rentatges"}
		case "rentament":
			return []string{"rentaments"}
		case "equilibratge":
			return []string{"equilibratges"}
		case "equilibrament":
			return []string{"equilibraments"}
		default:
			if strings.HasSuffix(lemma, "a") {
				return []string{lemma[:len(lemma)-1] + "es"}
			}
			return []string{lemma + "s"}
		}
	}

	matches := rule.Match(analyzeOpNames("Els equilibrats de les rodes"))
	require.NotEmpty(t, matches, "Els equilibrats de les rodes")

	matches = rule.Match(analyzeOpNames("El repicat i el rejuntat."))
	require.Equal(t, 2, len(matches), "El repicat i el rejuntat.")

	matches = rule.Match(analyzeOpNames("El procés de relligat dels llibres."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "relligadura", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "relligament", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "relligada", matches[0].GetSuggestedReplacements()[2])

	matches = rule.Match(analyzeOpNames("Els rentats de cervell."))
	require.Equal(t, 1, len(matches))
	// Without ConvertToGenderAndNumberFilter.Tag: plural surface synth only
	// (Java: "Les rentades" after gender/number filter).
	require.Contains(t, matches[0].GetSuggestedReplacements(), "rentades")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "rentatges")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "rentaments")
}

func TestReplaceOperationNamesRule_FailClosedWithoutPOS(t *testing.T) {
	rule := NewReplaceOperationNamesRule(nil)
	// Mid-sentence without det POS: fail closed
	matches := rule.Match(languagetool.AnalyzePlain("Cal vigilar el filtrat del vi"))
	require.Empty(t, matches)
	// Sentence-initial still matches via SENT_START prev
	matches = rule.Match(languagetool.AnalyzePlain("Assecat del braç del riu"))
	require.NotEmpty(t, matches)
}
