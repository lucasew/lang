package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/CatalanWordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalanWordRepeatRule_Rule(t *testing.T) {
	rule := NewCatalanWordRepeatRule(map[string]string{"repetition": "Repetició"})
	ok := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(s))), "ok %q", s)
	}
	bad := func(s string) {
		t.Helper()
		require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain(s))), "bad %q", s)
	}
	ok("Sempre pensa en en Joan.")
	ok("Els els portaré aviat.")
	ok("Maximilià I i Maria de Borgonya")
	ok("De la A a la z")
	ok("Entre I i II.")
	ok("fills de Sigebert I i Brunegilda")
	ok("del segle I i del segle II")
	ok("entre el capítol I i el II")
	ok("cada una una casa")
	ok("cada un un llibre")
	ok("Si no no es gaudeix.")
	ok("HUCHA-GANGA.ES es presenta.")
	ok("Ja fa, arreu arreu, més de quaranta anys.")
	ok("obrim inscripcions\U0001F44D\U0001F49A\U0001F332\U0001F332")
	ok("Anirem del punt A al punt B.")
	ok("La grip A a l'abril repunta.")
	ok("L'apartat A a la part final.")

	bad("Tots els els homes són iguals.")
	bad("Maximilià i i Maria de Borgonya")
}
