package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/CatalanWordRepeatRuleTest.java
import (
	"strings"
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
	// Java JLanguageTool + Catalan tagger/disambig assign _allow_repeat / LOC_ADV.
	// AnalyzePlain has no tagger — inject those tags for twin good cases (no surface invent).
	okTagged := func(s string, allowRepeatAtSecond bool, locAdvAtSecond bool) {
		t.Helper()
		sent := languagetool.AnalyzePlain(s)
		injectCAWordRepeatTags(sent, allowRepeatAtSecond, locAdvAtSecond)
		require.Equal(t, 0, len(rule.Match(sent)), "okTagged %q", s)
	}

	okTagged("Sempre pensa en en Joan.", true, false)
	okTagged("Els els portaré aviat.", true, false)
	okTagged("Maximilià I i Maria de Borgonya", true, false)
	okTagged("De la A a la z", true, false)
	okTagged("Entre I i II.", true, false)
	okTagged("fills de Sigebert I i Brunegilda", true, false)
	okTagged("del segle I i del segle II", true, false)
	okTagged("entre el capítol I i el II", true, false)
	okTagged("cada una una casa", true, false)
	okTagged("cada un un llibre", true, false)
	okTagged("Si no no es gaudeix.", true, false)
	okTagged("HUCHA-GANGA.ES es presenta.", true, false)
	okTagged("Ja fa, arreu arreu, més de quaranta anys.", false, true) // LOC_ADV
	// emoji / tree repeats: not isWord → no match without special tags
	ok("obrim inscripcions\U0001F44D\U0001F49A\U0001F332\U0001F332")
	okTagged("Anirem del punt A al punt B.", true, false)
	okTagged("La grip A a l'abril repunta.", true, false)
	okTagged("L'apartat A a la part final.", true, false)

	// incorrect: no special tags (fail closed) — same as Java without _allow_repeat
	bad("Tots els els homes són iguals.")
	bad("Maximilià i i Maria de Borgonya")
}

func TestCatalanWordRepeatRule_FailClosedWithoutTags(t *testing.T) {
	rule := NewCatalanWordRepeatRule(map[string]string{"repetition": "Repetició"})
	// Without Catalan disambig tags, "en en" is a repetition (no surface invent).
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Sempre pensa en en Joan."))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Els els portaré aviat."))))
}

func TestCatalanWordRepeatRule_LemmaIgnore(t *testing.T) {
	rule := NewCatalanWordRepeatRule(map[string]string{"repetition": "Repetició"})
	// Java hasLemma("Joan-Lluís Lluís") / "Chitty Chitty Bang Bang" on current token.
	sent := languagetool.AnalyzePlain("Lluís Lluís va escriure.")
	nws := sent.GetTokensWithoutWhitespace()
	// second Lluís at index of second content word
	for i, tok := range nws {
		if tok != nil && tok.GetToken() == "Lluís" && i > 1 {
			lemma := "Joan-Lluís Lluís"
			tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), nil, &lemma), "test")
			break
		}
	}
	require.Equal(t, 0, len(rule.Match(sent)), "lemma Joan-Lluís Lluís should ignore")
}

// injectCAWordRepeatTags marks the second token of the first equal-fold word pair
// with Java ignore tags (_allow_repeat and/or LOC_ADV).
func injectCAWordRepeatTags(sent *languagetool.AnalyzedSentence, allowRepeat, locAdv bool) {
	if sent == nil || (!allowRepeat && !locAdv) {
		return
	}
	nws := sent.GetTokensWithoutWhitespace()
	for i := 2; i < len(nws); i++ {
		if nws[i] == nil || nws[i-1] == nil {
			continue
		}
		a, b := nws[i-1].GetToken(), nws[i].GetToken()
		if !strings.EqualFold(a, b) {
			continue
		}
		// Found first repeated pair; tag current like Catalan disambig.
		if allowRepeat {
			tag := "_allow_repeat"
			nws[i].AddReading(languagetool.NewAnalyzedToken(b, &tag, nil), "test")
		}
		if locAdv {
			tag := "LOC_ADV"
			nws[i].AddReading(languagetool.NewAnalyzedToken(b, &tag, nil), "test")
		}
		return
	}
}
