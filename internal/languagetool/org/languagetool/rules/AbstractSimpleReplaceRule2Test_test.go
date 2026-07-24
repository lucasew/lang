package rules

// Twin of languagetool-core/src/test/java/org/languagetool/rules/AbstractSimpleReplaceRule2Test.java
import (
	"embed"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

//go:embed data/abstract_simple_replace2.txt
var asr2TestFS embed.FS

func newTestASR2(cs CaseSensitivity) *AbstractSimpleReplaceRule2 {
	f, err := asr2TestFS.Open("data/abstract_simple_replace2.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	r := &AbstractSimpleReplaceRule2{
		ID:              "ABSTRACT_TEST_RULE",
		Description:     "internal test rule",
		ShortMsg:        "internal test rule",
		MessageTemplate: "fake suggestion",
		CaseSens:        cs,
		LanguageCode:    "en",
	}
	if err := r.LoadSimpleReplaceRule2Data(f, "/xx/abstract_simple_replace2.txt"); err != nil {
		panic(err)
	}
	return r
}

func TestAbstractSimpleReplaceRule2_Rule(t *testing.T) {
	csRule := newTestASR2(CaseSensitive)
	require.Equal(t, 1, len(csRule.Match(languagetool.AnalyzePlain("But a propos"))))
	require.Equal(t, 0, len(csRule.Match(languagetool.AnalyzePlain("But A propos"))))
	require.Equal(t, 0, len(csRule.Match(languagetool.AnalyzePlain("A propos"))))
	require.Equal(t, 1, len(csRule.Match(languagetool.AnalyzePlain("a propos"))))
	require.Equal(t, 1, len(csRule.Match(languagetool.AnalyzePlain("A Pokemon"))))
	require.Equal(t, 0, len(csRule.Match(languagetool.AnalyzePlain("A pokemon"))))

	ciRule := newTestASR2(CaseInsensitive)
	require.Equal(t, 1, len(ciRule.Match(languagetool.AnalyzePlain("But a propos"))))
	require.Equal(t, 1, len(ciRule.Match(languagetool.AnalyzePlain("But A propos"))))
	require.Equal(t, 1, len(ciRule.Match(languagetool.AnalyzePlain("A propos"))))
	require.Equal(t, 1, len(ciRule.Match(languagetool.AnalyzePlain("a propos"))))
	require.Equal(t, 1, len(ciRule.Match(languagetool.AnalyzePlain("A Pokemon"))))
	require.Equal(t, 1, len(ciRule.Match(languagetool.AnalyzePlain("A pokemon"))))
}

// Java useSubRuleSpecificIds → SpecificIdRule(toId(getId()+"_"+original), desc.replace("$match", …)).
func TestAbstractSimpleReplaceRule2_SubRuleSpecificIDs(t *testing.T) {
	r := newTestASR2(CaseSensitive)
	r.SubRuleSpecificIDs = true
	r.Description = "Fix $match casing"
	ms := r.Match(languagetool.AnalyzePlain("But a propos"))
	require.Len(t, ms, 1)
	// Match rule is SpecificIdRule, not the base ABSTRACT_TEST_RULE
	idRule, ok := ms[0].GetRule().(*SpecificIdRule)
	require.True(t, ok, "match.GetRule must be SpecificIdRule when SubRuleSpecificIDs")
	require.Contains(t, idRule.GetID(), "ABSTRACT_TEST_RULE")
	require.NotEqual(t, "ABSTRACT_TEST_RULE", idRule.GetID())
	require.Equal(t, "Fix a propos casing", idRule.GetDescription())
}
