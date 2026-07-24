package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/SimpleReplaceAnglicismTest.java
// Without ConvertToGenderAndNumberFilter; assertions use surface replacements from replace_anglicism.txt.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceAnglicism_Rule(t *testing.T) {
	rule := NewSimpleReplaceAnglicism(nil)

	// correct (adapted forms / non-entries)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Això és un zombi molt perillós."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("El pòdcast d'avui ha estat molt interessant."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Hem rebut molts tiquets de suport."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("La bretxa digital és un problema greu."))))

	// graphic adaptations — surface only (no determiner rewrite)
	checkFirst := func(sentence, want string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q got %d", sentence, len(matches))
		require.Equal(t, want, matches[0].GetSuggestedReplacements()[0], "sentence %q", sentence)
	}
	checkFirst("El zombie era molt perillós.", "zombi")
	checkFirst("Els zombies eren molt perillosos.", "zombis")
	checkFirst("He sentit el podcast.", "pòdcast")
	checkFirst("He sentit els podcasts.", "pòdcasts")
	checkFirst("Tenim un ticket de suport obert.", "tiquet")
	checkFirst("Tenim dos tickets de suport oberts.", "tiquets")
	checkFirst("El troll va inundar el fòrum de missatges.", "trol")
	checkFirst("El footing és un esport popular.", "fúting")

	matches := rule.Match(languagetool.AnalyzePlain("Hi ha un canvi en l'snorkeling gratuït."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "esnòrquel", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "immersió lleugera", matches[0].GetSuggestedReplacements()[1])

	// unnecessary anglicisms
	matches = rule.Match(languagetool.AnalyzePlain("El spam és un problema al correu electrònic."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "correu brossa", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "contingut brossa", matches[0].GetSuggestedReplacements()[1])

	checkFirst("Hi ha un gap important entre els salaris.", "bretxa")

	matches = rule.Match(languagetool.AnalyzePlain("El cringe que em va fer aquell moment."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"angúnia", "vergonya", "incomoditat"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(languagetool.AnalyzePlain("Ens van programar uns briefings."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"sessions informatives", "reunions informatives", "reports", "brífings"},
		matches[0].GetSuggestedReplacements())
	// Surface match is the token only; Java gender filter may widen to include determiner.
	require.Equal(t, 22, matches[0].GetFromPos())
	require.Equal(t, 31, matches[0].GetToPos())

	// multiword — no gender filter
	matches = rule.Match(languagetool.AnalyzePlain("El vol era low cost."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"baix cost", "de baix cost", "barat", "barats"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, 11, matches[0].GetFromPos())
	require.Equal(t, 19, matches[0].GetToPos())

	onlineSuggs := []string{"en línia", "digital", "electrònic", "connectat", "per internet", "en remot", "en internet"}
	matches = rule.Match(languagetool.AnalyzePlain("Farem el seminari online."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, onlineSuggs, matches[0].GetSuggestedReplacements())

	matches = rule.Match(languagetool.AnalyzePlain("Farem el seminari on line."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 18, matches[0].GetFromPos())
	require.Equal(t, 25, matches[0].GetToPos())
	require.Equal(t, onlineSuggs, matches[0].GetSuggestedReplacements())

	matches = rule.Match(languagetool.AnalyzePlain("Farem el seminari on-line."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 18, matches[0].GetFromPos())
	require.Equal(t, 25, matches[0].GetToPos())
	require.Equal(t, onlineSuggs, matches[0].GetSuggestedReplacements())

	matches = rule.Match(languagetool.AnalyzePlain("Necessitem el know-how necessari per fer-ho."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 14, matches[0].GetFromPos())
	require.Equal(t, 22, matches[0].GetToPos())
	require.Equal(t, "saber fer", matches[0].GetSuggestedReplacements()[0])

	// sentence-start capitalization
	checkFirst("Zombie és el nom de la pel·lícula.", "Zombi")

	matches = rule.Match(languagetool.AnalyzePlain("El spam i el troll fan malbé les discussions en línia."))
	require.Equal(t, 2, len(matches))

	matches = rule.Match(languagetool.AnalyzePlain("La primera masterclass serà la de literatura barroca."))
	require.Equal(t, 1, len(matches))
	// surface only — no article/gender expansion
	require.Equal(t, []string{"classe magistral", "classes magistrals"}, matches[0].GetSuggestedReplacements())
}

func TestSimpleReplaceAnglicism_EnglishIgnoreException(t *testing.T) {
	rule := NewSimpleReplaceAnglicism(nil)
	// "spam" is in replace_anglicism; with adjacent _english_ignore_ should skip.
	sent := languagetool.AnalyzePlain("El spam is bad.")
	// tokens: SENT_START, El, spam, is, bad, .
	// mark spam + is with _english_ignore_ so startIndex and startIndex-1?
	// Java: startIndex > 1 && tokens[startIndex]._english_ignore_ && tokens[startIndex-1]._english_ignore_
	// For "spam" startIndex is spam; need prev also _english_ignore_
	// Use "The spam is bad" with The and spam tagged
	sent = languagetool.AnalyzePlain("The spam is bad.")
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		low := tok.GetToken()
		if low == "The" || low == "spam" || low == "is" || low == "bad" {
			pos := "_english_ignore_"
			tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
		}
	}
	matches := rule.Match(sent)
	require.Empty(t, matches, "english span should suppress anglicism match")

	// without tags, spam still matches
	matches = rule.Match(languagetool.AnalyzePlain("The spam is bad."))
	require.NotEmpty(t, matches)
}
