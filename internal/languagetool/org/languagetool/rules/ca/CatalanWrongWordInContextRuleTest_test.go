package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/CatalanWrongWordInContextRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalanWrongWordInContextRule_Rule(t *testing.T) {
	rule := NewCatalanWrongWordInContextRule(nil)
	assertN := func(s string, n int) {
		t.Helper()
		require.Equal(t, n, len(rule.Match(languagetool.AnalyzePlain(s))), "text %q", s)
	}
	assertN("Li va infringir un mal terrible.", 1)
	assertN("És un terreny abonat per als problemes.", 1)
	assertN("No li va cosir bé les betes.", 1)
	assertN("Sempre li seguia la beta.", 1)
	// pali / pal·li
	assertN("Sota els palis.", 1)
	assertN("Els pal·lis.", 0)
	assertN("El pal·li i el sànscrit.", 1)
	assertN("El pali i el sànscrit.", 0)
	assertN("Vam comprar xocolate de mànec.", 1)
	assertN("El pic de l'ocell.", 1)
}
