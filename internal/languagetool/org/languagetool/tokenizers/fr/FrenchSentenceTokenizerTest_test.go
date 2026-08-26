package fr

// Twin of languagetool-language-modules/fr/src/test/java/org/languagetool/tokenizers/fr/FrenchSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(French.getInstance()) — default paragraph mode (single-line breaks do not mark paragraphs).
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitFR mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitFR(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewFrenchSRXSentenceTokenizer()
	// match Java French SRX default paragraph mode (do NOT invent flags unless Java sets them)
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of FrenchSentenceTokenizerTest.testTokenize — all active cases, exact equality + size asserts.
func TestFrenchSentenceTokenizer_Tokenize(t *testing.T) {
	testSplitFR(t, "Je suis Chris.")
	testSplitFR(t, "Je suis Chris.")
	testSplitFR(t, "Je suis Chris ?!")
	testSplitFR(t, "Je suis Chris ?")
	testSplitFR(t, "Je suis      Chris ?")
	testSplitFR(t, "Je suis Chris...")
	testSplitFR(t, "Je suis Chris ...")
	testSplitFR(t, "Je suis Chris …")
	testSplitFR(t, "Votre nom: Chris !")
	testSplitFR(t, "Je suis (...) Chris")
	testSplitFR(t, "Je suis Chris (Christopher?).")
	testSplitFR(t, "Je suis Chris (Christopher ?).")
	testSplitFR(t, "Je suis Chris (Christopher ?!).")
	testSplitFR(t, "Je suis E. Macron de France.")
	testSplitFR(t, "J'ai beaucoup d'amis (Tom, Lisa, ...).")
	testSplitFR(t, "J'ai beaucoup d'amis (Tom, Lisa, ... ).")
	testSplitFR(t, "J'ai beaucoup d'amis (Tom, Lisa, …).")
	testSplitFR(t, "La fréquence des P.A. et le nombre de fibres recrutées.")
	testSplitFR(t, "Mrs. America est une mini-série américaine créée par Dahvi Waller, diffusée depuis le 15 avril 2020 sur le site de VOD Hulu et la chaîne FX.")
	testSplitFR(t, "Il travaille pour Tiffany & Co. à Paris.")
	testSplitFR(t, "J'ai beaucoup d'amis (Tom, Lisa, ...) et je suis populaire !")
	testSplitFR(t, "J'ai beaucoup d'amis (Tom, Lisa !) et je suis populaire !")
	testSplitFR(t, "Ph.D. est un groupe de musique britannique.")
	testSplitFR(t, "Google Inc. est une entreprise américaine")
	testSplitFR(t, "Le discours de E. Philippe devrait nous éclairer (un peu, beaucoup, ...?) sur ce qui nous attend.")
	testSplitFR(t, "Le discours de E. Philippe devrait nous éclairer (un peu, beaucoup, ...?) sur ce qui nous attend.")
	testSplitFR(t, "Op. cit., op. cit.")
	testSplitFR(t, "IVe siècle av. J.C. en architecture")
	testSplitFR(t, "IVe\u00a0siècle\u00a0av.\u00a0J.C.\u00a0en\u00a0architecture")
	testSplitFR(t, "IVe siècle av. J.-C. en architecture")
	testSplitFR(t, "sa mort le 19 août 14 apr. J.-C.")
	testSplitFR(t, "Je suis Chris.[4] ", "Je suis Chris.")
	testSplitFR(t, "Je suis Chris.[4]\u00a0", "Je suis Chris.")
	testSplitFR(t, "gaffa.org")
	testSplitFR(t, "Notice BnF de l'éd. Jean Marx.")
	testSplitFR(t, "L'Éducation nationale, impr. par ordre de la Convention nationale, Reprod. de l'éd. de : [Paris].")

	testSplitFR(t, "Le discours de E. Philippe devrait nous éclairer (un peu, beaucoup, …?) sur ce qui nous attend.")
	testSplitFR(t, "Le discours de E. Philippe devrait nous éclairer (un peu, beaucoup, … ?) sur ce qui nous attend.")
	testSplitFR(t, "Comment ça va … ?")

	// without nbsp
	testSplitFR(t, "« Le film était bien ? » ", "« Il était énorme ! ", "J'ai eu mal au ventre tellement je me suis marré ! »")
	testSplitFR(t, "Si « cf. » désigne l’abréviation de « confer »")
	// with nbsp
	testSplitFR(t, "« Le film était bien ? » ", "« Il était énorme ! ", "J'ai eu mal au ventre tellement je me suis marré ! »")
	testSplitFR(t, "Si « cf. » désigne l’abréviation de « confer »,")
	testSplitFR(t, "Ça ne sert à rien de me dire « Salut, comment ça va ? » si tu n'as rien d'autre à dire.")
	testSplitFR(t, "« Madame est dans sa chambre. » dit le serviteur.")
	testSplitFR(t, "« L'État, c'est moi ! » dit le roi.")

	tok := NewFrenchSRXSentenceTokenizer()
	require.Equal(t, 2, len(tok.Tokenize("Je suis Chris. Comment allez vous ?")))
	require.Equal(t, 2, len(tok.Tokenize("Je suis Chris?   Comment allez vous ???")))
	require.Equal(t, 2, len(tok.Tokenize("Je suis Chris ! Comment allez vous ???")))
	require.Equal(t, 2, len(tok.Tokenize("Je suis Chris ? Comment allez vous ???")))
	require.Equal(t, 2, len(tok.Tokenize("Je suis Chris. comment allez vous")))
	require.Equal(t, 2, len(tok.Tokenize("Je suis Chris (...). comment allez vous")))
	require.Equal(t, 2, len(tok.Tokenize("Je suis Chris (la la la …). comment allez vous")))
	require.Equal(t, 2, len(tok.Tokenize("Je suis Chris (CHRISTOPHER!). Comment allez vous")))
	require.Equal(t, 2, len(tok.Tokenize("Je suis Chris... Comment allez vous.")))

	testSplitFR(t, "Je ferai ça... un autre jour.")
	testSplitFR(t, "Qu'en dites-vous ?, demanda-t-il.")
	testSplitFR(t, "Qu'en dites-vous ? demanda-t-il.")
	testSplitFR(t, "Qu'en dites-vous ! demanda-t-il.")

	testSplitFR(t, "Première phrase.Deuxième phrase.")
}
