package jekavian

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/rules/sr/jekavian/SimpleGrammarJekavianReplaceRuleTest.java
// Java unit test body is empty (getMessage empty). Behavior matrix asserts Java-visible
// rule API outcomes and official header-only replace-grammar.txt load (empty wrongWords).

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestSimpleGrammarJekavianReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleGrammarJekavianReplaceRule(TestTools.getEnglishMessages())
	rule := NewSimpleGrammarJekavianReplaceRule(nil)

	// Metadata twin of SimpleGrammarJekavianReplaceRule overrides + AbstractSimpleReplaceRule ctor.
	require.Equal(t, SimpleGrammarJekavianReplaceRuleID, rule.GetID())
	require.Equal(t, "SR_JEKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE", rule.GetID())
	require.Equal(t, "Провера граматички погрешних ријечи или израза", rule.GetDescription())
	require.Equal(t, "Граматички погрешна ријеч тј. израз", rule.GetShort())
	require.Equal(t, "sr", rule.GetLocaleCode())
	require.True(t, rule.CheckLemmas)   // Java AbstractSimpleReplaceRule default
	require.True(t, rule.CaseSensitive) // Java isCaseSensitive default; SR does not override
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), rule.GetCategory().GetID())
	require.Equal(t, rules.ITSUncategorized, rule.GetLocQualityIssueType())

	// Official /sr/jekavian/replace-grammar.txt is header-only (comment lines).
	// Java SimpleReplaceDataLoader skips # lines → empty wrongWords map.
	require.NotNil(t, rule.GetWrongWords())
	require.Equal(t, 0, len(rule.GetWrongWords()),
		"jekavian replace-grammar.txt is header-only; wrongWords size must match Java empty load")
	require.Equal(t, SimpleGrammarJekavianReplaceGrammarPath, "/sr/jekavian/replace-grammar.txt")

	// getMessage exact Serbian quotes (U+201E „ and U+201C “) and join — Java override,
	// even though dictionary has no entries yet (test body empty upstream).
	wantMsg := "Не каже се „фоо“ него „бар, баз“."
	require.Equal(t, wantMsg, rule.GetMessage("фоо", []string{"бар", "баз"}))
	// Single replacement form used when matches fire:
	require.Equal(t, "Не каже се „вишљи“ него „виши“.", rule.GetMessage("вишљи", []string{"виши"}))

	// Match outcomes: empty dictionary → no matches on arbitrary Serbian text.
	// Java would behave the same with empty wrongWords.
	matches := rule.Match(languagetool.AnalyzePlain("Данас је диван дан."))
	require.Equal(t, 0, len(matches))
	matches = rule.Match(languagetool.AnalyzePlain("Син је вишљи од оца."))
	require.Equal(t, 0, len(matches), "header-only grammar dict must not invent ekavian keys")
	matches = rule.Match(languagetool.AnalyzePlain("У то оправдано сумљам."))
	require.Equal(t, 0, len(matches))
}

func TestSimpleGrammarJekavianReplaceRule_OfficialResourceByteIdentical(t *testing.T) {
	// Embedded data must be the official Java resource, byte-identical.
	embedded, err := grammarFS.ReadFile("data/replace-grammar.txt")
	require.NoError(t, err)

	// Resolve inspiration tree relative to this test file (repo layout).
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// .../internal/languagetool/org/languagetool/rules/sr/jekavian/this_test.go
	// → repo root is 7 parents up from jekavian/
	repoRoot := thisFile
	for i := 0; i < 8; i++ {
		repoRoot = filepath.Dir(repoRoot)
	}
	official := filepath.Join(repoRoot,
		"inspiration/languagetool/languagetool-language-modules/sr/src/main/resources/org/languagetool/rules/sr/jekavian/replace-grammar.txt")
	want, err := os.ReadFile(official)
	if err != nil {
		// Fallback: walk up until inspiration/ is found (workspace may differ).
		dir := filepath.Dir(thisFile)
		var found []byte
		for i := 0; i < 12; i++ {
			candidate := filepath.Join(dir,
				"inspiration/languagetool/languagetool-language-modules/sr/src/main/resources/org/languagetool/rules/sr/jekavian/replace-grammar.txt")
			b, e := os.ReadFile(candidate)
			if e == nil {
				found = b
				break
			}
			dir = filepath.Dir(dir)
		}
		require.NotNil(t, found, "could not locate official jekavian replace-grammar.txt under inspiration/")
		want = found
		err = nil
	}
	require.NoError(t, err)
	require.True(t, bytes.Equal(want, embedded),
		"embedded jekavian/data/replace-grammar.txt must be byte-identical to official Java resource")

	// Load via same loader Java uses → empty map for header-only file.
	m, err := rules.LoadSimpleReplaceWords(bytes.NewReader(embedded))
	require.NoError(t, err)
	require.Equal(t, 0, len(m))
}
