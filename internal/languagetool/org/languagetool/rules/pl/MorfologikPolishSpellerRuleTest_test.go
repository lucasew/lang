package pl

// Twin of languagetool-language-modules/pl/.../MorfologikPolishSpellerRuleTest.java
// Official resource: /pl/hunspell/pl_PL.dict (+ spelling.txt / ignore.txt / spelling_global via InitSpellersFromGetters).
// Analysis ports JLanguageTool(Polish).getAnalyzedSentence — PL tokenizer + PolishTagger + hybrid disambiguator.
// No map-inject invent; sole proof is real pl_PL.dict + real match outcomes.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	displ "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/pl"
	tagpl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pl"
	"github.com/stretchr/testify/require"
)

// requirePolishSpellerDict skips when official pl_PL.dict is not discoverable (no invent).
func requirePolishSpellerDict(t *testing.T) *MorfologikPolishSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(PolishSpellerDict) == "" {
		t.Skip("pl_PL.dict not discoverable (inspiration/.../pl/hunspell)")
	}
	r := NewMorfologikPolishSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// analyzePLSpeller ports JLanguageTool(new Polish()).getAnalyzedSentence.
// Wires PolishTagger + PolishHybridDisambiguator when official resources are present
// (required for ignore_numbers_compounds / hyphen adj compounds / MI-6 digit split).
func analyzePLSpeller(t *testing.T, text string) *languagetool.AnalyzedSentence {
	t.Helper()
	lt := languagetool.NewJLanguageTool("pl")
	if tagpl.DiscoverPolishPOSDict() != "" {
		tagpl.EnsureDefaultPolishTagger()
		if tg := tagpl.DefaultPolishTagger; tg != nil {
			lt.SentenceTagger = tg.Tag
			lt.TagWord = func(token string) []languagetool.TokenTag {
				readings := tg.Tag([]string{token})
				if len(readings) == 0 || readings[0] == nil {
					return nil
				}
				out := make([]languagetool.TokenTag, 0, 4)
				for _, at := range readings[0].GetReadings() {
					if at == nil {
						continue
					}
					pos, lemma := "", ""
					if p := at.GetPOSTag(); p != nil {
						pos = *p
					}
					if l := at.GetLemma(); l != nil {
						lemma = *l
					}
					if pos != "" || lemma != "" {
						out = append(out, languagetool.TokenTag{POS: pos, Lemma: lemma})
					}
				}
				return out
			}
		}
	}
	// Java createDefaultDisambiguator → PolishHybridDisambiguator (XML then multiwords).
	if displ.DiscoverPolishDisambiguationXML() != "" || displ.DiscoverPolishMultiwords() != "" {
		if hyb := displ.NewPolishHybridDisambiguator(); hyb != nil {
			lt.Disambiguator = hyb
		}
	}
	s := lt.GetAnalyzedSentence(text)
	require.NotNil(t, s)
	return s
}

// Port of MorfologikPolishSpellerRuleTest.testMorfologikSpeller
func TestMorfologikPolishSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requirePolishSpellerDict(t)
	require.Equal(t, MorfologikPolishSpellerRuleID, r.GetID())
	require.Equal(t, PolishSpellerDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_PL_PL", r.GetID())
	require.Equal(t, "/pl/hunspell/pl_PL.dict", r.GetFileName())

	// Hybrid disambiguation needed for 15-letni (ignore_numbers_compounds) and POS compounds.
	if tagpl.DiscoverPolishPOSDict() == "" {
		t.Skip("polish.dict not discoverable — needed for JLanguageTool(Polish) analysis path")
	}
	if displ.DiscoverPolishDisambiguationXML() == "" {
		t.Skip("pl/disambiguation.xml not discoverable — needed for ignore_numbers_compounds")
	}

	matchN := func(s string) int {
		t.Helper()
		ms, err := r.Match(analyzePLSpeller(t, s))
		require.NoError(t, err)
		return len(ms)
	}
	// correct sentences:
	require.Equal(t, 0, matchN("To jest test bez jakiegokolwiek błędu."))
	require.Equal(t, 0, matchN("Żółw na starość wydziela dziwną woń."))
	require.Equal(t, 0, matchN("Żółw na starość wydziela dziwną woń numer 1234."))

	require.Equal(t, 0, matchN("W MI-6 pracuje 15-letni agent."))
	// test for "LanguageTool":
	require.Equal(t, 0, matchN("LanguageTool jest świetny!"))

	// test for the ignored uppercase word "Gdym":
	require.Equal(t, 0, matchN("Gdym to zobaczył, zdębiałem."))

	require.Equal(t, 0, matchN(","))
	require.Equal(t, 0, matchN("123454"))

	// test for accentuated (foreign) words, such as "purée"
	require.Equal(t, 0, matchN("Roztrzepać widelcem jajka razem z purée z dyni."))

	// compound word with ignored part "techniczno"
	require.Equal(t, 0, matchN("Bogactwo nie rośnie proporcjonalnie do jej rozwoju techniczno-terytorialnego."))

	// compound word with one of the compound prefixes:
	require.Equal(t, 0, matchN("Antypostmodernistyczna batalia hiperfilozofów"))
	// compound words: "trzynastobitowy", "zgniłożółty"
	require.Equal(t, 0, matchN("Trzynastobitowe przystawki w kolorze zgniłożółtym"))

	// compound word that failed test when CamelCase was on:
	require.Equal(t, 0, matchN("Nie spotkałem tutaj Saint-Exupérych"))

	// incorrect sentences:
	ms, err := r.Match(analyzePLSpeller(t, "Zolw"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 4, ms[0].GetToPos())
	require.Equal(t, "Żółw", ms[0].GetSuggestedReplacements()[0])

	ms, err = r.Match(analyzePLSpeller(t, "😂 Zolw"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, 3, ms[0].GetFromPos())
	require.Equal(t, 7, ms[0].GetToPos())
	require.Equal(t, "Żółw", ms[0].GetSuggestedReplacements()[0])

	ms, err = r.Match(analyzePLSpeller(t, "😂😂 Zolw"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, 5, ms[0].GetFromPos())
	require.Equal(t, 9, ms[0].GetToPos())
	require.Equal(t, "Żółw", ms[0].GetSuggestedReplacements()[0])

	require.Equal(t, 1, matchN("aõh"))

	// tokenizing on prefixes niby- and quasi-
	require.Equal(t, 0, matchN("Niby-artysta spotkał się z quasi-opiekunem i niby-Francuzem."))

	pruned, err := r.Match(analyzePLSpeller(t, "Clarkem"))
	require.NoError(t, err)
	require.Len(t, pruned, 1)
	sugs := pruned[0].GetSuggestedReplacements()
	require.Len(t, sugs, 6)
	require.Equal(t, "Clarke", sugs[0])
	require.Equal(t, "Clarkiem", sugs[1])
	require.Equal(t, "Ciarkę", sugs[2])
	require.Equal(t, "Clarkom", sugs[3])
	require.Equal(t, "Czarkę", sugs[4])

	// with default ignore-camel-case == yes this was not catched at all
	require.Equal(t, 1, matchN("Saint-Exupérymu"))

	// There should be a match, this is not a prefix!
	require.Equal(t, 1, matchN("premoc"))

	// "0" instead "o"...
	require.Equal(t, 1, matchN("dziwneg0"))
}
