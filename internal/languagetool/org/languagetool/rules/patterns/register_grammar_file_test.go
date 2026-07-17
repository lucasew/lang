package patterns_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestRegisterGrammarFile_SoftEN(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	// patterns → rules → languagetool → org → languagetool → internal → module root (6)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../.."))
	path := filepath.Join(root, "testdata/grammar/en-soft.xml")
	lt := languagetool.NewJLanguageTool("en")
	n, err := patterns.RegisterGrammarFile(lt, path, "en")
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)
	m := lt.Check("Well, your welcome to try.")
	found := false
	for _, x := range m {
		if x.RuleID == "EN_SOFT_YOUR_YOU_RE" {
			found = true
			require.Contains(t, x.Message, "you're welcome")
			if len(x.Suggestions) > 0 {
				require.Contains(t, x.Suggestions, "you're welcome")
			}
		}
	}
	require.True(t, found, "%+v", m)
}

func TestRegisterGrammarXML_Inline(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	xml := `<rules lang="en"><category id="G"><rule id="X"><pattern><token>foo</token><token>bar</token></pattern><message>bad <suggestion>baz</suggestion></message></rule></category></rules>`
	n, err := patterns.RegisterGrammarXML(lt, xml, "inline", "en")
	require.NoError(t, err)
	require.Equal(t, 1, n)
	m := lt.Check("say foo bar now")
	require.NotEmpty(t, m)
}

func TestRegisterGrammarFile_SoftDE(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../.."))
	path := filepath.Join(root, "testdata/grammar/de-soft.xml")
	lt := languagetool.NewJLanguageTool("de")
	n, err := patterns.RegisterGrammarFile(lt, path, "de")
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)
	m := lt.Check("Ich denke das es stimmt.")
	found := false
	for _, x := range m {
		if x.RuleID == "DE_SOFT_DAS_DASS" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}

func TestRegisterSoftGrammarDir_RU_SV_DA(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../.."))
	dir := filepath.Join(root, "testdata/grammar")

	for _, tc := range []struct {
		lang, text, ruleID string
	}{
		{"ru", "пошел в в магазин", "RU_SOFT_V_V"},
		{"sv", "men dom är här", "SV_SOFT_DE_DOM"},
		{"da", "en del af af det", "DA_SOFT_AF_AF"},
		{"uk", "пішов в в магазин", "UK_SOFT_V_V"},
		{"ca", "va anar a a casa", "CA_SOFT_A_A"},
		{"gl", "casa de de pedra", "GL_SOFT_DE_DE"},
		{"sk", "a a b", "SK_SOFT_A_A"},
		{"el", "και και άλλο", "EL_SOFT_KAI_KAI"},
		{"ar", "كتاب في في البيت", "AR_SOFT_FI_FI"},
		{"ro", "casa de de piatră", "RO_SOFT_DE_DE"},
		{"br", "ha ha bras", "BR_SOFT_HA_HA"},
		{"fa", "کتاب و و دفتر", "FA_SOFT_VA_VA"},
		{"ga", "agus agus eile", "GA_SOFT_AGUS_AGUS"},
		{"sl", "in in drugo", "SL_SOFT_IN_IN"},
		{"km", "and and more", "KM_SOFT_AND_AND"},
		{"be", "і і слова", "BE_SOFT_I_I"},
		{"eo", "kaj kaj pli", "EO_SOFT_KAJ_KAJ"},
		{"is", "og og meira", "IS_SOFT_OG_OG"},
		{"ja", "to to more", "JA_SOFT_TO_TO"},
		{"lt", "ir ir daugiau", "LT_SOFT_IR_IR"},
		{"zh", "的 的 字", "ZH_SOFT_DE_DE"},
		{"sr", "i i drugo", "SR_SOFT_I_I"},
		{"ta", "um um more", "TA_SOFT_UM_UM"},
		{"tl", "at at pa", "TL_SOFT_AT_AT"},
	} {
		t.Run(tc.lang, func(t *testing.T) {
			lt := languagetool.NewJLanguageTool(tc.lang)
			n, err := patterns.RegisterSoftGrammarDir(lt, dir, tc.lang)
			require.NoError(t, err)
			require.GreaterOrEqual(t, n, 1)
			m := lt.Check(tc.text)
			found := false
			for _, x := range m {
				if x.RuleID == tc.ruleID {
					found = true
					require.Equal(t, "GRAMMAR", x.CategoryID)
					require.Equal(t, "Grammar", x.CategoryName)
					require.Equal(t, "grammar", x.IssueType)
				}
			}
			require.True(t, found, "%+v", m)
		})
	}
}
