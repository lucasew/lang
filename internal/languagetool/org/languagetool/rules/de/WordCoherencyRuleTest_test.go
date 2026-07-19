package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WordCoherencyRuleTest.java
// Production loads coherency.txt pairs only (no invent suffixes). Inflected forms
// match via lemmas — Java uses GermanTagger; tests attach the same lemmas the
// morph dict would (fixture when german.dict is absent).
import (
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// twinCoherencyLemmas: surface (lower) → lemma for WordCoherencyRuleTest inputs.
// Mirrors GermanTagger base forms for those surfaces (not a production invent path).
var twinCoherencyLemmas = map[string]string{
	"aufwendig": "aufwendig", "aufwändig": "aufwändig",
	"aufwendige": "aufwendig", "aufwändige": "aufwändig",
	"aufwendiger": "aufwendig", "aufwändiger": "aufwändig",
	"aufwendigste": "aufwendig", "aufwändigste": "aufwändig",
	"delfin": "delfin", "delphin": "delphin",
	"delfins": "delfin", "delphine": "delphin",
	"essentiell": "essentiell", "essenziell": "essenziell",
	"essentieller": "essentiell", "essenzielles": "essenziell",
	"differential": "differential", "differenzial": "differenzial",
	"differentials": "differential", "differenzials": "differenzial",
	"facette": "facette", "fassette": "fassette",
	"facetten": "facetten", "fassetten": "fassetten",
	"joghurt": "joghurt", "jogurt": "jogurt",
	// plurals lemma to singular base pairs in coherency.txt (Java morph)
	"joghurts": "joghurt", "jogurts": "jogurt",
	"ketchup": "ketchup", "ketschup": "ketschup",
	"ketchups": "ketchup", "ketschups": "ketschup",
	"kommuniqué": "kommuniqué", "kommunikee": "kommunikee",
	"kommuniqués": "kommuniqué", "kommunikees": "kommunikee",
	"necessaire": "necessaire", "nessessär": "nessessär",
	"necessaires": "necessaire", "nessessärs": "nessessär",
	"orthographie": "orthographie", "orthografie": "orthografie",
	"substantiell": "substantiell", "substanziell": "substanziell",
	"substantieller": "substantiell", "substanzielles": "substanziell",
	"thunfisch": "thunfisch", "tunfisch": "tunfisch",
	// plural forms also listed as pairs in coherency.txt
	"thunfische": "thunfische", "tunfische": "tunfische",
	"xylophon": "xylophon", "xylofon": "xylofon",
	"xylophone": "xylophon", "xylofone": "xylofon",
	"selbständig": "selbständig", "selbstständig": "selbstständig",
	"selbständiges": "selbständig", "selbstständiger": "selbstständig",
	"bahnhofsplatz": "bahnhofsplatz", "bahnhofplatz": "bahnhofplatz",
}

func analyzeDECoherency(s string) []*languagetool.AnalyzedSentence {
	// Sentence-local positions like AnalyzeTextLocal, with lemmas on content tokens.
	if s == "" {
		return nil
	}
	var parts []string
	start := 0
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '.' || r == '!' || r == '?' {
			if r == '.' && i+1 < len(runes) {
				n := runes[i+1]
				if (n >= 'a' && n <= 'z') || (n >= '0' && n <= '9') {
					continue
				}
			}
			end := i + 1
			if end < len(runes) && (runes[end] == ' ' || runes[end] == '\n' || runes[end] == '\u00A0') {
				if runes[end] == '\n' && end+1 < len(runes) && runes[end+1] == '\n' {
					end++
					if end < len(runes) && runes[end] == '\n' {
						end++
					}
				} else if runes[end] == ' ' || runes[end] == '\u00A0' {
					end++
				} else if runes[end] == '\n' {
					end++
				}
			}
			parts = append(parts, string(runes[start:end]))
			start = end
			i = end - 1
		}
	}
	if start < len(runes) {
		parts = append(parts, string(runes[start:]))
	}
	out := make([]*languagetool.AnalyzedSentence, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, languagetool.AnalyzeWithTagger(p, deCoherencyTagWord))
	}
	return out
}

var deCoherencyTaggerCached = openDiscoveredGermanTagger(DiscoverGermanResourceDir())

func deCoherencyTagWord(tok string) []languagetool.TokenTag {
	if tok == "" {
		return nil
	}
	// Prefer real GermanTagger when resources exist (Java path).
	if deCoherencyTaggerCached != nil {
		if rd := deCoherencyTaggerCached.Lookup(tok); rd != nil {
			var tags []languagetool.TokenTag
			for _, r := range rd.GetReadings() {
				if r == nil {
					continue
				}
				tt := languagetool.TokenTag{}
				if p := r.GetPOSTag(); p != nil {
					tt.POS = *p
				}
				if l := r.GetLemma(); l != nil {
					tt.Lemma = *l
				}
				if tt.POS != "" || tt.Lemma != "" {
					tags = append(tags, tt)
				}
			}
			if len(tags) > 0 {
				return tags
			}
		}
	}
	// Fixture lemmas for twin-test surfaces when morph dict is missing.
	key := strings.ToLower(tok)
	key = strings.TrimFunc(key, func(r rune) bool {
		return !unicode.IsLetter(r) && r != 'ä' && r != 'ö' && r != 'ü' && r != 'ß' && r != 'é' && r != 'á'
	})
	if lem, ok := twinCoherencyLemmas[key]; ok {
		return []languagetool.TokenTag{{Lemma: lem}}
	}
	return nil
}

// Untagged AnalyzeTextLocal must not invent suffix maps (production expand=false).
func TestWordCoherencyRule_NoInventWithoutLemma(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	// Plain analysis has no lemmas → inflected surfaces not in coherency.txt must not match.
	ms := rule.MatchList(languagetool.AnalyzeTextLocal("Das ist aufwendiger, aber nicht zu aufwändig."))
	require.Equal(t, 0, len(ms), "fail-closed without invent expand / lemmas")
}

func TestWordCoherencyRule_Rule(t *testing.T) {
	assertGood := func(s string) {
		t.Helper()
		rule := NewWordCoherencyRule(nil)
		require.Equal(t, 0, len(rule.MatchList(analyzeDECoherency(s))), "good: %q", s)
	}
	assertError := func(s string, expectedSuggestion ...string) {
		t.Helper()
		rule := NewWordCoherencyRule(nil)
		matches := rule.MatchList(analyzeDECoherency(s))
		require.Equal(t, 1, len(matches), "error: %q got %d", s, len(matches))
		if len(expectedSuggestion) > 0 {
			require.Equal(t, expectedSuggestion[0], matches[0].GetSuggestedReplacements()[0], "sugg for %q", s)
		}
	}

	assertGood("Das ist aufwendig, aber nicht zu aufwendig.")
	assertGood("Das ist aufwendig. Aber nicht zu aufwendig.")
	assertGood("Das ist aufwändig, aber nicht zu aufwändig.")
	assertGood("Das ist aufwändig. Aber nicht zu aufwändig.")

	assertError("Das ist aufwendig. Aufwändig ist das.", "Aufwendig")
	assertError("Das ist aufwendig, aber nicht zu aufwändig.")
	assertError("Das ist aufwendig. Aber nicht zu aufwändig.")
	assertError("Das ist aufwendiger, aber nicht zu aufwändig.")
	assertError("Das ist aufwendiger. Aber nicht zu aufwändig.")
	assertError("Das ist aufwändig, aber nicht zu aufwendig.")
	assertError("Das ist aufwändig. Aber nicht zu aufwendig.")
	assertError("Das ist aufwändiger, aber nicht zu aufwendig.")
	assertError("Das ist aufwändiger. Aber nicht zu aufwendig.")
	assertError("Delfin und Delphin")
	assertError("Delfins und Delphine")
	assertError("essentiell und essenziell")
	assertError("essentieller und essenzielles")
	assertError("Differential und Differenzial")
	assertError("Differentials und Differenzials")
	assertError("Facette und Fassette")
	assertError("Facetten und Fassetten")
	assertError("Joghurt und Jogurt")
	assertError("Joghurts und Jogurt")
	assertError("Joghurt und Jogurts")
	assertError("Joghurts und Jogurts")
	assertError("Ketchup und Ketschup")
	assertError("Ketchups und Ketschups")
	assertError("Kommuniqué und Kommunikee")
	assertError("Kommuniqués und Kommunikees")
	assertError("Necessaire und Nessessär")
	assertError("Necessaires und Nessessärs")
	assertError("Orthographie und Orthografie")
	assertError("substantiell und substanziell")
	assertError("substantieller und substanzielles")
	assertError("Thunfisch und Tunfisch")
	assertError("Thunfische und Tunfische")
	assertError("Xylophon und Xylofon")
	assertError("Xylophone und Xylofone")
	assertError("selbständig und selbstständig")
	assertError("selbständiges und selbstständiger")
	assertError("Bahnhofsplatz und Bahnhofplatz")

	rule := NewWordCoherencyRule(nil)
	matches1 := rule.MatchList(analyzeDECoherency("Eine aufwendige Untersuchung. Oder ist sie aufwändig?"))
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 43, matches1[0].GetFromPos())
	require.Equal(t, 52, matches1[0].GetToPos())
	require.Equal(t, []string{"aufwendig"}, matches1[0].GetSuggestedReplacements())

	matches2 := rule.MatchList(analyzeDECoherency("Eine aufwendige Untersuchung. Oder ist sie noch aufwändiger?"))
	require.Equal(t, 1, len(matches2))
	require.Equal(t, 48, matches2[0].GetFromPos())
	require.Equal(t, 59, matches2[0].GetToPos())
	require.Equal(t, []string{"aufwendiger"}, matches2[0].GetSuggestedReplacements())
}

func TestWordCoherencyRule_CallIndependence(t *testing.T) {
	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(NewWordCoherencyRule(nil).MatchList(analyzeDECoherency(s))))
	}
	assertGood("Das ist aufwendig.")
	assertGood("Aber nicht zu aufwändig.")
}

func TestWordCoherencyRule_MatchPosition(t *testing.T) {
	matches := NewWordCoherencyRule(nil).MatchList(analyzeDECoherency("Das ist aufwendig. Aber nicht zu aufwändig"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 33, matches[0].GetFromPos())
	require.Equal(t, 42, matches[0].GetToPos())
}

func TestWordCoherencyRule_RuleCompleteTexts(t *testing.T) {
	check := func(s string) int {
		return len(NewWordCoherencyRule(nil).MatchList(analyzeDECoherency(s)))
	}
	require.Equal(t, 0, check("Das ist aufwändig. Aber hallo. Es ist wirklich aufwändig."))
	require.Equal(t, 1, check("Das ist aufwendig. Aber hallo. Es ist wirklich aufwändig."))
	require.Equal(t, 1, check("Das ist aufwändig. Aber hallo. Es ist wirklich aufwendig."))
	require.Equal(t, 0, check("Das ist aufwendig. Aber hallo. Es ist wirklich aufwendiger als so."))
	require.Equal(t, 1, check("Das ist aufwendig. Aber hallo. Es ist wirklich aufwändiger als so."))
	require.Equal(t, 1, check("Das ist aufwändig. Aber hallo. Es ist wirklich aufwendiger als so."))
	require.Equal(t, 1, check("Das ist das aufwändigste. Aber hallo. Es ist wirklich aufwendiger als so."))
	require.Equal(t, 1, check("Das ist das aufwändigste. Aber hallo. Es ist wirklich aufwendig."))
	// cross-paragraph: AnalyzeTextLocal may not split on \n\n alone — use AnalyzeTextDemo if needed
	require.Equal(t, 1, check("Das ist das aufwändigste.\n\nAber hallo. Es ist wirklich aufwendig."))
}


func TestWordCoherencyRule_CategoryAndMinParagraph(t *testing.T) {
	r := NewWordCoherencyRule(nil)
	require.NotNil(t, r.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), r.GetCategory().GetID())
	require.Equal(t, -1, r.MinToCheckParagraph())
}
