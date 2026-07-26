package jekavian

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/rules/sr/jekavian/SimpleStyleJekavianReplaceRuleTest.java
// Java unit test body is empty (getMessage empty). Behavior matrix asserts Java-visible
// rule API outcomes and official header-only replace-style.txt load (empty wrongWords).

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

func TestSimpleStyleJekavianReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleStyleJekavianReplaceRule(TestTools.getEnglishMessages())
	rule := NewSimpleStyleJekavianReplaceRule(nil)

	// Metadata twin of SimpleStyleJekavianReplaceRule overrides + AbstractSimpleReplaceRule ctor.
	require.Equal(t, SimpleStyleJekavianReplaceRuleID, rule.GetID())
	require.Equal(t, "SR_JEKAVIAN_SIMPLE_STYLE_REPLACE_RULE", rule.GetID())
	require.Equal(t, "Провера стилски лоших ријечи или израза", rule.GetDescription())
	require.Equal(t, "Стилски лоша ријеч тј. израз", rule.GetShort())
	require.Equal(t, "sr", rule.GetLocaleCode())
	require.True(t, rule.CheckLemmas)   // Java AbstractSimpleReplaceRule default
	require.True(t, rule.CaseSensitive) // Java isCaseSensitive default; SR does not override
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), rule.GetCategory().GetID())
	require.Equal(t, rules.ITSUncategorized, rule.GetLocQualityIssueType())

	// Official /sr/jekavian/replace-style.txt is header-only (comment lines).
	// Java SimpleReplaceDataLoader skips # lines → empty wrongWords map.
	require.NotNil(t, rule.GetWrongWords())
	require.Equal(t, 0, len(rule.GetWrongWords()),
		"jekavian replace-style.txt is header-only; wrongWords size must match Java empty load")
	require.Equal(t, SimpleStyleJekavianReplaceStylePath, "/sr/jekavian/replace-style.txt")

	// getMessage exact Serbian quotes (U+201E „ and U+201C “) and join — Java override,
	// even though dictionary has no entries yet (test body empty upstream).
	// Java: "Умјесто израза „" + tokenStr + "“ било би боље да користите: "
	//        + String.join(", ", replacements) + "."
	wantMsg := "Умјесто израза „фоо“ било би боље да користите: бар, баз."
	require.Equal(t, wantMsg, rule.GetMessage("фоо", []string{"бар", "баз"}))
	// Single replacement form used when matches fire:
	require.Equal(t, "Умјесто израза „дебела“ било би боље да користите: елегантно попуњена.",
		rule.GetMessage("дебела", []string{"елегантно попуњена"}))

	// Match outcomes: empty dictionary → no matches on arbitrary Serbian text.
	// Java would behave the same with empty wrongWords.
	// Do not invent ekavian style dictionary entries into jekavian path.
	matches := rule.Match(languagetool.AnalyzePlain("Данас је диван дан."))
	require.Equal(t, 0, len(matches))
	matches = rule.Match(languagetool.AnalyzePlain("Она је дебела."))
	require.Equal(t, 0, len(matches), "header-only style dict must not invent ekavian keys")
	matches = rule.Match(languagetool.AnalyzePlain("Купио је компјутер."))
	require.Equal(t, 0, len(matches))
}

func TestSimpleStyleJekavianReplaceRule_OfficialResourceByteIdentical(t *testing.T) {
	// Embedded data must be the official Java resource, byte-identical.
	embedded, err := styleFS.ReadFile("data/replace-style.txt")
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
		"inspiration/languagetool/languagetool-language-modules/sr/src/main/resources/org/languagetool/rules/sr/jekavian/replace-style.txt")
	want, err := os.ReadFile(official)
	if err != nil {
		// Fallback: walk up until inspiration/ is found (workspace may differ).
		dir := filepath.Dir(thisFile)
		var found []byte
		for i := 0; i < 12; i++ {
			candidate := filepath.Join(dir,
				"inspiration/languagetool/languagetool-language-modules/sr/src/main/resources/org/languagetool/rules/sr/jekavian/replace-style.txt")
			b, e := os.ReadFile(candidate)
			if e == nil {
				found = b
				break
			}
			dir = filepath.Dir(dir)
		}
		require.NotNil(t, found, "could not locate official jekavian replace-style.txt under inspiration/")
		want = found
		err = nil
	}
	require.NoError(t, err)
	require.True(t, bytes.Equal(want, embedded),
		"embedded jekavian/data/replace-style.txt must be byte-identical to official Java resource")

	// Load via same loader Java uses → empty map for header-only file.
	m, err := rules.LoadSimpleReplaceWords(bytes.NewReader(embedded))
	require.NoError(t, err)
	require.Equal(t, 0, len(m))
}
