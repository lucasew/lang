package de

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestPotentialCompoundFilter_FailClosedWithoutSpeller(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewPotentialCompoundFilter()
	// Without speller hook: fail-closed → hyphenated only (do not invent joined valid).
	s := f.Suggestions("Haus", "tür")
	require.Equal(t, []string{"Haus-Tür"}, s)
}

func TestPotentialCompoundFilter_MisspelledHyphenatedOnly(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewPotentialCompoundFilter()
	f.JoinedIsValid = func(string) bool { return false }
	require.Equal(t, []string{"Haus-Tür"}, f.Suggestions("Haus", "Tür"))
	require.Equal(t, []string{"Xyzzy-Qqqq"}, f.Suggestions("Xyzzy", "qqqq"))
}

func TestPotentialCompoundFilter_ValidJoinedIncludesJoined(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewPotentialCompoundFilter()
	f.JoinedIsValid = func(string) bool { return true }
	s := f.Suggestions("Haus", "tür")
	require.Equal(t, []string{"Haustür"}, s)
}

func TestPotentialCompoundFilter_ValidLongAddsHyphenatedFirst(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewPotentialCompoundFilter()
	f.JoinedIsValid = func(string) bool { return true }
	// joined "Superlangerwortxyzab" length: need UTF-16 length > 20
	// part1 "Superlangerwort" (15) + part2 "xyzabc" (6) = 21 > 20
	part1 := "Superlangerwort" // 15
	part2 := "xyzabc"         // 6 → joined 21
	s := f.Suggestions(part1, part2)
	require.Equal(t, 21, utf16LenDE("Superlangerwortxyzabc"))
	require.Equal(t, []string{"Superlangerwort-Xyzabc", "Superlangerwortxyzabc"}, s)

	// boundary: length == 20 → no hyphenated first
	// 14 + 6 = 20
	s20 := f.Suggestions("Superlangerwor", "xyzabc")
	require.Equal(t, 20, utf16LenDE("Superlangerworxyzabc"))
	require.Equal(t, []string{"Superlangerworxyzabc"}, s20)
}

func TestPotentialCompoundFilter_UTF16LengthBoundary(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewPotentialCompoundFilter()
	f.JoinedIsValid = func(string) bool { return true }
	// Non-BMP: one Go rune = 2 UTF-16 units (surrogate pair).
	// "A" + 10 × 😀 (U+1F600) = 1 + 20 = 21 UTF-16 units > 20
	emoji := strings.Repeat("😀", 10)
	// part2 emoji-only: IsAllUppercase true (no lower letter) → keep as-is.
	// part1 "A": IsAllUppercase true → keep. joined UTF-16 length = 1 + 20 = 21.
	joined := "A" + emoji
	require.Equal(t, 21, utf16LenDE(joined), "Java String.length of A+10 emoji")
	s := f.Suggestions("A", emoji)
	require.Equal(t, []string{"A-" + emoji, joined}, s)
}

func TestPotentialCompoundFilter_CapitalizationMatrix(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewPotentialCompoundFilter()
	f.JoinedIsValid = func(string) bool { return false } // hyphenated only — assert casing on hyphen form

	// Normal: capitalize first of each, part2 lower when joined
	require.Equal(t, []string{"Haus-Tür"}, f.Suggestions("haus", "tür"))
	require.Equal(t, []string{"Haus-Tür"}, f.Suggestions("Haus", "Tür"))

	// All-upper: keep as-is
	require.Equal(t, []string{"HAUS-TÜR"}, f.Suggestions("HAUS", "TÜR"))

	// Mixed case: keep as-is (isMixedCase)
	require.Equal(t, []string{"iPod-Case"}, f.Suggestions("iPod", "Case"))
	// part2 mixed keeps original for both lower and cap forms → joined iPodcAsE when valid
	f.JoinedIsValid = func(string) bool { return true }
	require.Equal(t, []string{"iPodcAsE"}, f.Suggestions("iPod", "cAsE"))
	// part1 all-upper, part2 normal
	f.JoinedIsValid = func(string) bool { return false }
	require.Equal(t, []string{"HAUS-Tür"}, f.Suggestions("HAUS", "tür"))
	// part1 normal, part2 all-upper
	require.Equal(t, []string{"Haus-TÜR"}, f.Suggestions("haus", "TÜR"))
}

func TestPotentialCompoundFilter_PreservesType(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewPotentialCompoundFilter()
	f.JoinedIsValid = func(string) bool { return false }
	for _, typ := range []rules.RuleMatchType{
		rules.RuleMatchTypeHint,
		rules.RuleMatchTypeUnknownWord,
		rules.RuleMatchTypeOther,
	} {
		m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 8, "msg")
		m.SetType(typ)
		m.IssueType = "should-not-be-copied"
		out := f.AcceptRuleMatch(m, map[string]string{"part1": "Haus", "part2": "tür"}, 0, nil, nil)
		require.NotNil(t, out)
		require.Equal(t, typ, out.GetType(), "setType(match.getType())")
		require.Equal(t, []string{"Haus-Tür"}, out.GetSuggestedReplacements())
		require.Equal(t, "msg", out.GetMessage())
		require.Equal(t, "", out.IssueType)
	}
}

func TestPotentialCompoundFilter_AcceptRuleMatchShortMessage(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewPotentialCompoundFilter()
	f.JoinedIsValid = func(string) bool { return true }
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 1, 5, "long msg")
	m.ShortMessage = "short"
	out := f.AcceptRuleMatch(m, map[string]string{"part1": "Abend", "part2": "essen"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, "short", out.ShortMessage)
	require.Equal(t, []string{"Abendessen"}, out.GetSuggestedReplacements())
}

func TestPotentialCompoundFilter_WithInjectedSpeller(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewPotentialCompoundFilter()
	f.JoinedIsValid = func(string) bool { return true }
	s2 := f.Suggestions("Haus", "tür")
	require.Contains(t, s2, "Haustür")
	f.JoinedIsValid = func(string) bool { return false }
	s3 := f.Suggestions("Haus", "Tür")
	require.Equal(t, []string{"Haus-Tür"}, s3)
}

func TestPotentialCompoundFilter_WithDEDict(t *testing.T) {
	ClearGermanFilterSpeller()
	t.Cleanup(ClearGermanFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	dict := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/de_DE.dict")
	if !WireGermanFilterSpeller(dict) {
		t.Skipf("de_DE.dict not openable: %s", dict)
	}
	f := NewPotentialCompoundFilter()
	// No override: GermanSpellerRule.Match on mini sentence (Java default spelling rule).
	// Unknown junk → hyphenated only
	s := f.Suggestions("Xyzzy", "qqqq")
	require.Equal(t, []string{"Xyzzy-Qqqq"}, s)
	// Abendessen is typically in de_DE as a compound form — if known, suggest joined
	s2 := f.Suggestions("Abend", "essen")
	if FilterDictIsMisspelled("Abendessen") {
		require.Equal(t, []string{"Abend-Essen"}, s2)
	} else {
		require.Contains(t, s2, "Abendessen")
	}
	// Match path equivalence: dict misspell ⇒ Match non-empty for junk
	require.False(t, f.joinedValid("Xyzzyqqqq"))
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 4, "msg")
	m.SetType(rules.RuleMatchTypeUnknownWord)
	out := f.AcceptRuleMatch(m, map[string]string{"part1": "Xyzzy", "part2": "qqqq"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Xyzzy-Qqqq"}, out.GetSuggestedReplacements())
	require.Equal(t, rules.RuleMatchTypeUnknownWord, out.GetType())
}

func TestPotentialCompoundFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.de.PotentialCompoundFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.de.PotentialCompoundFilter"))
}

func TestPotentialCompoundFilter_BuildSentenceHasSentStart(t *testing.T) {
	sent := buildPotentialCompoundSentence("Haustür")
	require.NotNil(t, sent)
	toks := sent.GetTokens()
	require.GreaterOrEqual(t, len(toks), 2)
	require.True(t, toks[0].IsSentenceStart())
	require.Equal(t, "Haustür", toks[1].GetToken())
}
