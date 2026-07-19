package ca

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikCatalanSpellerRule(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	// Java MorfologikCatalanSpellerRule.getId / getFileName
	require.Equal(t, "MORFOLOGIK_RULE_CA_ES", MorfologikCatalanSpellerRuleID)
	require.Equal(t, "/ca/ca-ES_spelling.dict", CatalanSpellerDict)
	require.Equal(t, MorfologikCatalanSpellerRuleID, r.GetID())
	require.Equal(t, CatalanSpellerDict, r.GetFileName())
	// Java setIgnoreTaggedWords()
	require.True(t, r.IgnoreTaggedWords)
}

func TestMorfologikCatalanSpellerRule_IgnoreTagged(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(CatalanSpellerDict, 1)
	sp.AddWord("casa")
	r := NewMorfologikCatalanSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	sent := languagetool.AnalyzePlain("xyzzy")
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		pos := "N"
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
		require.True(t, tok.IsTagged())
	}
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestOrderCatalanSuggestions_DropCapitalizedDup(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	// Java: word.equals(word.toLowerCase()) && isCapitalizedWord(replacement)
	// && replacements.contains(replacement.toLowerCase())
	got := r.orderCatalanSuggestions([]string{"casa", "Casa", "caso"}, "caza")
	require.Equal(t, []string{"casa", "caso"}, got)
}

func TestOrderCatalanSuggestions_DropComo(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	got := r.orderCatalanSuggestions([]string{"como", "casa"}, "caza")
	require.Equal(t, []string{"casa"}, got)
}

func TestOrderCatalanSuggestions_Inalambric(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	got := r.orderCatalanSuggestions([]string{"inalàmbric", "casa"}, "inalambric")
	require.Equal(t, []string{"sense fils", "sense fil", "sense cables", "autònom"}, got)
}

func TestOrderCatalanSuggestions_PrefixAmbEspai(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	got := r.orderCatalanSuggestions([]string{"anti virus", "casa"}, "antivirus")
	require.Equal(t, []string{"casa"}, got)
}

func TestOrderCatalanSuggestions_ParticulaFront(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	got := r.orderCatalanSuggestions([]string{"zzz", "de casa"}, "decasa")
	require.Equal(t, []string{"de casa", "zzz"}, got)
}

func TestAdditionalTopCatalan_CamelCase(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(CatalanSpellerDict, 1)
	sp.AddWord("Guarda")
	sp.AddWord("Chuva")
	r := NewMorfologikCatalanSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	require.Equal(t, []string{"Guarda Chuva"}, r.additionalTopCatalanSuggestions("GuardaChuva"))
}

func TestAdditionalTopCatalan_DigitSplit(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	// without TagPOS fail-closed
	require.Empty(t, r.additionalTopCatalanSuggestions("casa2"))
	// short word not in SPLIT_DIGITS unless in list
	r.TagPOS = func(w string) []string {
		if w == "casa" || w == "de" {
			return []string{"NCFS000"}
		}
		return nil
	}
	require.Equal(t, []string{"casa 2"}, r.additionalTopCatalanSuggestions("casa2"))
	require.Equal(t, []string{"de 2"}, r.additionalTopCatalanSuggestions("de2"))
	// short non-list: "ab2" with tag still blocked (len ≤ 2 and not in list)
	r.TagPOS = func(w string) []string {
		if w == "ab" {
			return []string{"N"}
		}
		return nil
	}
	require.Empty(t, r.additionalTopCatalanSuggestions("ab2"))
}

func TestFindSuggestionCA_ApostrofInici(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	// without TagPOS fail-closed
	require.Empty(t, r.apostropheHyphenTopSuggestion("lamic"))
	r.TagPOS = func(w string) []string {
		// "amic" matches APOSTROF_INICI_NOM_SING stem (vowel + ≥3 chars)
		if strings.EqualFold(w, "amic") {
			return []string{"NCMS000"} // N..[SN].* → NOM_SING
		}
		return nil
	}
	got := r.apostropheHyphenTopSuggestion("lamic")
	require.Equal(t, "l'amic", got)
}

func TestFindSuggestionCA_GerundiHyphen(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	r.TagPOS = func(w string) []string {
		// group1 "cantan" + addStr "t" → "cantant" tagged as gerund V.G.*
		if w == "cantant" {
			return []string{"VMG0000"}
		}
		return nil
	}
	got := r.apostropheHyphenTopSuggestion("cantanhi")
	require.Equal(t, "cantant-hi", got)
}

func TestFindSuggestionMultiplePronouns(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	// without TagPOS
	require.Empty(t, r.findSuggestionMultiplePronouns("anarsen"))
	r.TagPOS = func(w string) []string {
		if w == "anar" {
			return []string{"VMN0000"} // V.N.* infinitive / also V.[NGM]
		}
		return nil
	}
	got := r.findSuggestionMultiplePronouns("anarsen")
	// verb "anar" + TransformDarrere("sen", "anar") — sen is in PronomsDarrere as "sen"
	require.NotEmpty(t, got)
	require.True(t, strings.HasPrefix(got, "anar"), "got %q", got)
}

func TestWireCatalanSpellerTagPOS(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	WireCatalanSpellerTagPOS(r, func(token string) []languagetool.TokenTag {
		if token == "casa" {
			return []languagetool.TokenTag{{POS: "NCFS000"}}
		}
		return nil
	})
	require.True(t, r.isTaggedCA("casa"))
	require.False(t, r.isTaggedCA("xyzzy"))
	require.Equal(t, []string{"casa 2"}, r.additionalTopCatalanSuggestions("casa2"))
}

func TestCatalanTokenizeNewWordsFalse(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	require.True(t, r.DisableTokenizeNewWords)
}

func TestCatalanUseInOffice(t *testing.T) {
	require.True(t, NewMorfologikCatalanSpellerRule().UseInOffice())
}
