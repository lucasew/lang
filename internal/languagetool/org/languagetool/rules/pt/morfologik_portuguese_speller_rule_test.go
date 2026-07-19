package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

func TestMorfologikPortugueseSpellerRule_PathsAndIDs(t *testing.T) {
	// Java MorfologikPortugueseSpellerRule.getDictFilename / dictFilepath / getId
	require.Equal(t, "/pt/spelling/pt-PT-90.dict", PortuguesePTDict)
	require.Equal(t, "/pt/spelling/pt-BR.dict", PortugueseBRDict)
	require.Equal(t, "MORFOLOGIK_RULE_PT_PT", MorfologikPortuguesePTSpellerRuleID)
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR", MorfologikPortugueseBRSpellerRuleID)

	pt := NewMorfologikPortugalPortugueseSpellerRule()
	require.Equal(t, MorfologikPortuguesePTSpellerRuleID, pt.GetID())
	require.Equal(t, PortuguesePTDict, pt.GetFileName())
	require.Equal(t, "pt-PT", pt.VariantCode)
	// Java getRuleMatches skips _english_ignore_
	require.NotNil(t, pt.SkipTokenFn)

	br := NewMorfologikBrazilianPortugueseSpellerRule()
	require.Equal(t, MorfologikPortugueseBRSpellerRuleID, br.GetID())
	require.Equal(t, PortugueseBRDict, br.GetFileName())
	require.Equal(t, "pt-BR", br.VariantCode)
	require.NotNil(t, br.SkipTokenFn)
}

func TestMorfologikPortugueseSpellerRule_EnglishIgnorePOS(t *testing.T) {
	r := NewMorfologikPortugalPortugueseSpellerRule()
	r.Speller.AddWord("ola")
	require.NotNil(t, r.SkipTokenFn)

	eng := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("something", strPtr("_english_ignore_"), nil),
	)
	require.True(t, r.SkipTokenFn(eng))
	require.False(t, r.SkipTokenFn(
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("xyzzy", nil, nil)),
	))

	sent := languagetool.AnalyzePlain("ola xyzzy")
	// without ignore: xyzzy flagged
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, m)

	toks := sent.GetTokensWithoutWhitespace()
	var content []*languagetool.AnalyzedTokenReadings
	for _, tok := range toks {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		content = append(content, tok)
	}
	require.GreaterOrEqual(t, len(content), 2)
	content[1].AddReading(languagetool.NewAnalyzedToken(content[1].GetToken(), strPtr("_english_ignore_"), nil), "test")
	m, err = r.Match(sent)
	require.NoError(t, err)
	for _, match := range m {
		from, to := match.GetFromPos(), match.GetToPos()
		text := sent.GetText()
		if from >= 0 && to <= len(text) && from < to {
			require.NotEqual(t, content[1].GetToken(), text[from:to])
		}
	}
}

func TestMorfologikPortugueseSpellerRule_LoadsPTRootWordLists(t *testing.T) {
	// Java getIgnoreFileName → "pt/ignore.txt" (resource root, not hunspell/)
	p := spelling.DiscoverLangHunspellWordList("pt", "ignore.txt")
	if p == "" {
		t.Skip("pt/ignore.txt not discoverable")
	}
	require.Contains(t, p, "ignore.txt")
	// should not require hunspell/ segment
	require.NotContains(t, p, "hunspell")

	r := NewMorfologikPortugalPortugueseSpellerRule()
	// official sanity entry in ignore.txt
	require.True(t, r.IgnoreWord("ignorewordoogaboogatest"),
		"expected ignorewordoogaboogatest from pt/ignore.txt")
	// prohibit.txt sanity entry
	if spelling.DiscoverLangHunspellWordList("pt", "prohibit.txt") != "" {
		require.True(t, r.IsProhibited("prohibitwordoogaboogatest") || !r.AcceptWord("prohibitwordoogaboogatest"),
			"expected prohibitwordoogaboogatest from pt/prohibit.txt")
	}
}

func strPtr(s string) *string { return &s }
