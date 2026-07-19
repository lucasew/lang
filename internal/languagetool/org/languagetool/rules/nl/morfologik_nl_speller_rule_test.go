package nl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

func TestMorfologikDutchSpellerRule(t *testing.T) {
	r := NewMorfologikDutchSpellerRule()
	// Java MorfologikDutchSpellerRule.getId / getFileName
	require.Equal(t, "MORFOLOGIK_RULE_NL_NL", MorfologikDutchSpellerRuleID)
	require.Equal(t, "/nl/spelling/nl_NL.dict", DutchSpellerDict)
	require.Equal(t, MorfologikDutchSpellerRuleID, r.GetID())
	require.Equal(t, DutchSpellerDict, r.GetFileName())
	// Java ignorePotentiallyMisspelledWord → CompoundAcceptor.acceptCompound
	require.NotNil(t, r.IgnorePotentiallyMisspelledWordFn)
	// empty acceptor (no lists) rejects; KnownWords path accepts
	require.False(t, r.IgnorePotentiallyMisspelledWord("onbekendwoordxyz"))
	DefaultCompoundAcceptor.KnownWords["testcompoundwoord"] = struct{}{}
	t.Cleanup(func() {
		delete(DefaultCompoundAcceptor.KnownWords, "testcompoundwoord")
	})
	require.True(t, r.IgnorePotentiallyMisspelledWord("testcompoundwoord"))
	// Java getRuleMatches skips _english_ignore_
	require.NotNil(t, r.SkipTokenFn)
}

func TestMorfologikDutchSpellerRule_EnglishIgnorePOS(t *testing.T) {
	r := NewMorfologikDutchSpellerRule()
	// inject tiny dict so AcceptWord would flag unknown words
	r.Speller.AddWord("hallo")
	require.NotNil(t, r.SkipTokenFn)

	eng := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("something", strPtr("_english_ignore_"), nil),
	)
	require.True(t, r.SkipTokenFn(eng))
	require.False(t, r.SkipTokenFn(
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("xyzzy", nil, nil)),
	))

	sent := languagetool.AnalyzePlain("hallo xyzzy")
	toks := sent.GetTokensWithoutWhitespace()
	var content []*languagetool.AnalyzedTokenReadings
	for _, tok := range toks {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		content = append(content, tok)
	}
	require.GreaterOrEqual(t, len(content), 2)
	// without ignore tag, xyzzy is flagged
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, m, "xyzzy should be misspelled without ignore tag")

	// with _english_ignore_, no match on that token
	content[1].AddReading(languagetool.NewAnalyzedToken(content[1].GetToken(), strPtr("_english_ignore_"), nil), "test")
	require.True(t, content[1].HasPosTag("_english_ignore_"))
	m, err = r.Match(sent)
	require.NoError(t, err)
	for _, match := range m {
		require.NotEqual(t, content[1].GetToken(), matchCoveredToken(match, sent))
	}
}

func TestMorfologikDutchSpellerRule_LoadsSpellingDirWordLists(t *testing.T) {
	// Java getIgnoreFileName → /nl/spelling/ignore.txt (not hunspell/)
	p := spelling.DiscoverLangHunspellWordList("nl", "ignore.txt")
	if p == "" {
		t.Skip("nl/spelling/ignore.txt not discoverable")
	}
	require.Contains(t, p, "spelling")
	r := NewMorfologikDutchSpellerRule()
	// ApplyDefault runs in NewMorfologikSpellerRule and should load nl/spelling lists.
	// Single-token ignore line "Abra" (hyphenated lines become MultiWordIgnore via tokenizer).
	require.True(t, r.IsInIgnoredSet("Abra") || r.IgnoreWord("Abra"),
		"expected single-token ignore entry Abra from nl/spelling/ignore.txt")
	// Hyphenated multi-token lines still registered as MultiWordIgnore phrases.
	require.NotEmpty(t, r.MultiWordIgnore)
}

func matchCoveredToken(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
	if m == nil || sent == nil {
		return ""
	}
	text := sent.GetText()
	from, to := m.GetFromPos(), m.GetToPos()
	if from < 0 || to > len(text) || from >= to {
		return ""
	}
	return text[from:to]
}

func strPtr(s string) *string { return &s }
