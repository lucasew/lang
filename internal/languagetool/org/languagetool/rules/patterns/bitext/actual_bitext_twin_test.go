package bitext

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/commandline"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/bitext"
	"github.com/stretchr/testify/require"
)

func findPath(rel string) string {
	wd, _ := os.Getwd()
	for dir := wd; ; dir = filepath.Dir(dir) {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		if filepath.Dir(dir) == dir {
			return ""
		}
	}
}

// Twin of ToolsTest.testBitextCheck (ACTUAL false-friend en→pl spans).
// Java: Tools.getBitextRules(en, pl) + Tools.checkBitext with full POS analysis.
func TestToolsBitextCheck_ACTUAL(t *testing.T) {
	ffPath := findPath(filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
		"org", "languagetool", "rules", "false-friends.xml"))
	if ffPath == "" {
		t.Skip("false-friends.xml missing")
	}
	plDict := commandline.DiscoverLanguagePOSDict(nil, "pl")
	if plDict == "" {
		t.Skip("PL POS dict required for inflected aktualny")
	}

	rules, err := LoadGetBitextRulesFromPaths("en", "pl", "", ffPath, "")
	require.NoError(t, err)
	// Filter to ACTUAL only (avoid SameTranslation etc. on short pairs)
	var actual []bitext.BitextRule
	for _, r := range rules {
		if r != nil && r.GetID() == "ACTUAL" {
			actual = append(actual, r)
		}
	}
	require.NotEmpty(t, actual, "ACTUAL from getBitextRules")

	enLT := languagetool.NewJLanguageTool("en")
	plLT := languagetool.NewJLanguageTool("pl")
	require.True(t, languagetool.RegisterBinaryPOSTagger(plLT, plDict))
	if enDict := commandline.DiscoverEnglishPOSDict(nil); enDict != "" {
		_ = languagetool.RegisterBinaryPOSTagger(enLT, enDict)
	}

	// Case 1: Java Tools.checkBitext uses getAnalyzedSentence (not multi-sentence Analyze).
	srcText := "This is not actual."
	trgText := "To nie jest aktualne."
	src := enLT.GetAnalyzedSentence(srcText)
	trg := plLT.GetAnalyzedSentence(trgText)
	require.NotNil(t, src)
	require.NotNil(t, trg)
	ms := bitext.CheckBitextAnalyzed(src, trg, trgText, actual)
	require.Len(t, ms, 1)
	require.Equal(t, "ACTUAL", ms[0].RuleID)
	require.Equal(t, 12, ms[0].FromPos)
	require.Equal(t, 20, ms[0].ToPos)

	// Case 2: multi-clause string — Java getAnalyzedSentence does not SRX-split.
	srcText2 := "A sentence. This is not actual."
	trgText2 := "Zdanie. To nie jest aktualne."
	src2 := enLT.GetAnalyzedSentence(srcText2)
	trg2 := plLT.GetAnalyzedSentence(trgText2)
	require.NotNil(t, src2)
	require.NotNil(t, trg2)
	require.Equal(t, srcText2, src2.GetText())
	require.Equal(t, trgText2, trg2.GetText())
	ms2 := bitext.CheckBitextAnalyzed(src2, trg2, trgText2, actual)
	require.Len(t, ms2, 1)
	require.Equal(t, "ACTUAL", ms2[0].RuleID)
	require.Equal(t, 20, ms2[0].FromPos)
	require.Equal(t, 28, ms2[0].ToPos)

	// Case 3: Java ToolsTest matches3 span 25-33
	srcText3 := "A new sentence. This is not actual."
	trgText3 := "Nowa zdanie. To nie jest aktualne."
	src3 := enLT.GetAnalyzedSentence(srcText3)
	trg3 := plLT.GetAnalyzedSentence(trgText3)
	ms3 := bitext.CheckBitextAnalyzed(src3, trg3, trgText3, actual)
	require.Len(t, ms3, 1)
	require.Equal(t, "ACTUAL", ms3[0].RuleID)
	require.Equal(t, 25, ms3[0].FromPos)
	require.Equal(t, 33, ms3[0].ToPos)
}

func TestGetBitextRules_IncludesBuiltins(t *testing.T) {
	rules, err := GetBitextRules("en", "pl", "", "", "")
	require.NoError(t, err)
	ids := map[string]bool{}
	for _, r := range rules {
		ids[r.GetID()] = true
	}
	// Java BitextRule.getRelevantRules IDs
	require.True(t, ids["TRANSLATION_LENGTH"])
	require.True(t, ids["SAME_TRANSLATION"])
	require.True(t, ids["DIFFERENT_PUNCTUATION"])
}

// Full Tools.checkBitext path: getBitextRules + getAnalyzedSentence + checkAnalyzedSentence + bitext.
func TestToolsCheckBitext_WithLanguageTools_ACTUAL(t *testing.T) {
	ffPath := findPath(filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
		"org", "languagetool", "rules", "false-friends.xml"))
	if ffPath == "" {
		t.Skip("false-friends.xml missing")
	}
	plDict := commandline.DiscoverLanguagePOSDict(nil, "pl")
	if plDict == "" {
		t.Skip("PL POS dict required")
	}
	rules, err := LoadGetBitextRulesFromPaths("en", "pl", "", ffPath, "")
	require.NoError(t, err)
	// Keep ACTUAL only so monolingual noise does not hide the twin
	var actual []bitext.BitextRule
	for _, r := range rules {
		if r != nil && r.GetID() == "ACTUAL" {
			actual = append(actual, r)
		}
	}
	require.NotEmpty(t, actual)

	enLT := languagetool.NewJLanguageTool("en")
	plLT := languagetool.NewJLanguageTool("pl")
	require.True(t, languagetool.RegisterBinaryPOSTagger(plLT, plDict))
	if enDict := commandline.DiscoverEnglishPOSDict(nil); enDict != "" {
		_ = languagetool.RegisterBinaryPOSTagger(enLT, enDict)
	}
	// Register a mono rule that would fire on target — ensure bitext still present
	plLT.AddRuleChecker("MONO_NOISE", func(s *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		return nil // no noise
	})

	ms := bitext.CheckBitextWithLanguageTools(
		"This is not actual.",
		"To nie jest aktualne.",
		enLT, plLT, actual,
	)
	require.NotEmpty(t, ms)
	found := false
	for _, m := range ms {
		if m.RuleID == "ACTUAL" {
			found = true
			require.Equal(t, 12, m.FromPos)
			require.Equal(t, 20, m.ToPos)
		}
	}
	require.True(t, found, "%+v", ms)
}
