package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJLanguageTool_CheckWordRepeat(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddChecker(SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Empty(t, lt.Check("This is fine."))
	m := lt.Check("This is is wrong.")
	require.Len(t, m, 1)
	require.Equal(t, "WORD_REPEAT_RULE", m[0].RuleID)
	require.Greater(t, m[0].ToPos, m[0].FromPos)
}

func TestJLanguageTool_CheckCancel(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddChecker(SimpleWordRepeatChecker(""))
	lt.Cancelled = func() bool { return true }
	require.Empty(t, lt.Check("is is"))
}

func TestJLanguageTool_UnknownWords(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.SetListUnknownWords(true)
	lt.IsKnownWord = KnownWordSet("This", "is", "a", "text")
	_ = lt.Check("This is a xyzzy text")
	require.Equal(t, []string{"xyzzy"}, lt.GetUnknownWords())
}

func TestCleanOverlappingLocalMatches(t *testing.T) {
	// non-overlap preserved (juxtaposed)
	require.Len(t, CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 0, ToPos: 2, Priority: 1},
		{FromPos: 3, ToPos: 5, Priority: 1},
	}), 2)
	// higher priority wins
	got := CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 0, ToPos: 4, RuleID: "a", Priority: 1},
		{FromPos: 1, ToPos: 3, RuleID: "b", Priority: 5},
	})
	require.Len(t, got, 1)
	require.Equal(t, "b", got[0].RuleID)

	// equal priority → longest span (Java)
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 0, ToPos: 4, RuleID: "short", Priority: 1},
		{FromPos: 1, ToPos: 20, RuleID: "long", Priority: 1},
	})
	require.Len(t, got, 1)
	require.Equal(t, "long", got[0].RuleID)

	// picky demotion: non-picky wins over picky at same base priority
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 0, ToPos: 10, RuleID: "picky", Priority: 1, IsPicky: true},
		{FromPos: 2, ToPos: 5, RuleID: "plain", Priority: 1},
	})
	require.Len(t, got, 1)
	require.Equal(t, "plain", got[0].RuleID)

	// three non-overlapping survive
	require.Len(t, CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 0, ToPos: 2, Priority: 1},
		{FromPos: 3, ToPos: 5, Priority: 1},
		{FromPos: 6, ToPos: 8, Priority: 1},
	}), 3)

	// both punctuation-only: prefer IncludedInErrorsCorrectedAllAtOnce (Java)
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{
			FromPos: 0, ToPos: 5, RuleID: "plain", Priority: 5,
			OriginalErrorStr: "Hallo", Suggestions: []string{"Hallo,"},
		},
		{
			FromPos: 0, ToPos: 5, RuleID: "allAtOnce", Priority: 1,
			OriginalErrorStr:                   "Hallo",
			Suggestions:                        []string{"Hallo!"},
			IncludedInErrorsCorrectedAllAtOnce: true,
		},
	})
	require.Len(t, got, 1)
	require.Equal(t, "allAtOnce", got[0].RuleID)

	// without OriginalErrorStr and without SentenceText: not punctuation-only → higher base priority wins
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 0, ToPos: 5, RuleID: "high", Priority: 5, Suggestions: []string{"Hallo,"}},
		{
			FromPos: 0, ToPos: 5, RuleID: "allAtOnce", Priority: 1,
			Suggestions:                        []string{"Hallo!"},
			IncludedInErrorsCorrectedAllAtOnce: true,
		},
	})
	require.Len(t, got, 1)
	require.Equal(t, "high", got[0].RuleID)

	// Java isPunctuationOnlyChange: OriginalErrorStr empty → sentence.substring(from,to)
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{
			FromPos: 0, ToPos: 5, RuleID: "plain", Priority: 5,
			SentenceText: "Hallo", FromPosSentence: 0, ToPosSentence: 5,
			Suggestions: []string{"Hallo,"},
		},
		{
			FromPos: 0, ToPos: 5, RuleID: "allAtOnce", Priority: 1,
			SentenceText: "Hallo", FromPosSentence: 0, ToPosSentence: 5,
			Suggestions:                        []string{"Hallo!"},
			IncludedInErrorsCorrectedAllAtOnce: true,
		},
	})
	require.Len(t, got, 1)
	require.Equal(t, "allAtOnce", got[0].RuleID)

	// Sentence fallback via document FromPos/ToPos when sentence positions unset (Java second branch)
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{
			FromPos: 0, ToPos: 5, RuleID: "plain", Priority: 5,
			SentenceText: "Hallo", Suggestions: []string{"Hallo,"},
		},
		{
			FromPos: 0, ToPos: 5, RuleID: "allAtOnce", Priority: 1,
			SentenceText:                       "Hallo",
			Suggestions:                        []string{"Hallo!"},
			IncludedInErrorsCorrectedAllAtOnce: true,
		},
	})
	require.Len(t, got, 1)
	require.Equal(t, "allAtOnce", got[0].RuleID)

	// HidePremiumMatches: premium demoted to MinInt32 → non-premium wins on overlap
	got = CleanOverlappingLocalMatchesOpts([]LocalMatch{
		{FromPos: 0, ToPos: 10, RuleID: "P2_PREMIUM_RULE", Priority: 10},
		{FromPos: 2, ToPos: 5, RuleID: "P1_RULE", Priority: 1},
	}, CleanOverlapOpts{HidePremiumMatches: true})
	require.Len(t, got, 1)
	require.Equal(t, "P1_RULE", got[0].RuleID)

	// Explicit IsPremium without PREMIUM in id
	got = CleanOverlappingLocalMatchesOpts([]LocalMatch{
		{FromPos: 0, ToPos: 10, RuleID: "SECRET", Priority: 10, IsPremium: true},
		{FromPos: 2, ToPos: 5, RuleID: "OPEN", Priority: 1},
	}, CleanOverlapOpts{HidePremiumMatches: true})
	require.Len(t, got, 1)
	require.Equal(t, "OPEN", got[0].RuleID)

	// DefaultPremium.IsPremiumRule (Java Premium.get().isPremiumRule) without IsPremium flag
	prevPremium := DefaultPremium
	DefaultPremium = premiumRuleIDs{ids: map[string]bool{"SECRET_VIA_REGISTRY": true}}
	t.Cleanup(func() { DefaultPremium = prevPremium })
	got = CleanOverlappingLocalMatchesOpts([]LocalMatch{
		{FromPos: 0, ToPos: 10, RuleID: "SECRET_VIA_REGISTRY", Priority: 10},
		{FromPos: 2, ToPos: 5, RuleID: "OPEN", Priority: 1},
	}, CleanOverlapOpts{HidePremiumMatches: true})
	require.Len(t, got, 1)
	require.Equal(t, "OPEN", got[0].RuleID)

	// Without hide, premium keeps higher priority
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 0, ToPos: 10, RuleID: "P2_PREMIUM_RULE", Priority: 10},
		{FromPos: 2, ToPos: 5, RuleID: "P1_RULE", Priority: 1},
	})
	require.Len(t, got, 1)
	require.Equal(t, "P2_PREMIUM_RULE", got[0].RuleID)

	// Java CleanOverlappingFilter: juxtaposed comma dup-suggestion treated as overlap
	// (prev ends with ",", cur starts with ", "; FromPos == prev.ToPos).
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 5, ToPos: 10, RuleID: "COMMA_LOW_PRIORITY", Priority: 1, Suggestions: []string{"right,"}},
		{FromPos: 10, ToPos: 15, RuleID: "COMMA_HIGH_PRIORITY", Priority: 10, Suggestions: []string{", left"}},
	})
	require.Len(t, got, 1)
	require.Equal(t, "COMMA_HIGH_PRIORITY", got[0].RuleID)

	// High priority first: still keep high when dup-sug with low
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 5, ToPos: 10, RuleID: "COMMA_HIGH_PRIORITY", Priority: 10, Suggestions: []string{"right,"}},
		{FromPos: 10, ToPos: 15, RuleID: "COMMA_LOW_PRIORITY", Priority: 1, Suggestions: []string{", left"}},
	})
	require.Len(t, got, 1)
	require.Equal(t, "COMMA_HIGH_PRIORITY", got[0].RuleID)

	// Equal priority: take last (Java currentPriority++)
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 5, ToPos: 10, RuleID: "COMMA_LOW_PRIORITY2", Priority: 1, Suggestions: []string{"right,"}},
		{FromPos: 10, ToPos: 15, RuleID: "COMMA_LOW_PRIORITY", Priority: 1, Suggestions: []string{", left"}},
	})
	require.Len(t, got, 1)
	require.Equal(t, "COMMA_LOW_PRIORITY", got[0].RuleID)

	// Java: multi-word suggestion share token across gap FromPos == ToPos+1
	// ("of the" + "the provisions" → treat as overlap; higher prio wins).
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 5, ToPos: 10, RuleID: "MISSING_THE_HIGH_PRIORITY", Priority: 10, Suggestions: []string{"of the"}},
		{FromPos: 11, ToPos: 15, RuleID: "MISSING_THE_LOW_PRIORITY", Priority: 1, Suggestions: []string{"the provisions"}},
	})
	require.Len(t, got, 1)
	require.Equal(t, "MISSING_THE_HIGH_PRIORITY", got[0].RuleID)

	// Juxtaposed without dup-sug pattern → both kept
	got = CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 0, ToPos: 4, RuleID: "A", Priority: 1, Suggestions: []string{"foo"}},
		{FromPos: 4, ToPos: 8, RuleID: "B", Priority: 1, Suggestions: []string{"bar"}},
	})
	require.Len(t, got, 2)
}

// premiumRuleIDs is a test Premium that marks listed rule IDs as premium.
type premiumRuleIDs struct {
	ids map[string]bool
}

func (p premiumRuleIDs) IsPremiumRule(ruleID string) bool {
	return p.ids[ruleID]
}

func TestJLanguageTool_AvsAnAndCorrect(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	src := "This is an test."
	m := lt.Check(src)
	require.NotEmpty(t, m)
	require.Equal(t, "This is a test.", CorrectTextFromLocalMatches(src, m))
}

func TestJLanguageTool_DisableRule(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.DisableRule("WORD_REPEAT_RULE")
	require.Empty(t, lt.Check("is is"))
	lt.EnableRule("WORD_REPEAT_RULE")
	require.NotEmpty(t, lt.Check("is is"))
}

// Java sentence.substring(from,to) is UTF-16; multi-byte German must not corrupt.
func TestOriginalSurface_UTF16MultiByte(t *testing.T) {
	m := LocalMatch{
		SentenceText: "Größe,",
		FromPos:      0,
		ToPos:        5, // "Größe" in UTF-16 units
	}
	require.Equal(t, "Größe", m.OriginalSurface())

	m2 := LocalMatch{
		SentenceText: "Größe,",
		FromPos:      0,
		ToPos:        6, // "Größe,"
	}
	require.Equal(t, "Größe,", m2.OriginalSurface())

	// Explicit OriginalErrorStr wins
	m3 := LocalMatch{OriginalErrorStr: "explicit", SentenceText: "Größe,", FromPos: 0, ToPos: 5}
	require.Equal(t, "explicit", m3.OriginalSurface())

	// Sentence positions preferred
	m4 := LocalMatch{
		SentenceText:    "Größe,",
		FromPos:         99,
		ToPos:           104,
		FromPosSentence: 0,
		ToPosSentence:   5,
	}
	require.Equal(t, "Größe", m4.OriginalSurface())
}

func TestCheck_AdjustLocalMatchPosLineColumn(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	// Two sentences so second match has non-zero document offset + line/column from SentenceData.
	ms := lt.Check("Hello. this this")
	require.NotEmpty(t, ms)
	// find word-repeat match
	var found *LocalMatch
	for i := range ms {
		if ms[i].RuleID == "WORD_REPEAT_RULE" {
			found = &ms[i]
			break
		}
	}
	require.NotNil(t, found)
	// sentence-relative positions preserved
	require.GreaterOrEqual(t, found.FromPosSentence, 0)
	require.Greater(t, found.ToPosSentence, found.FromPosSentence)
	// document offset should include first sentence
	require.Greater(t, found.FromPos, found.FromPosSentence)
	// line/column set by AdjustLocalMatchPos (not left at zero accidentally without running adjust)
	require.GreaterOrEqual(t, found.Column, 0)
	require.GreaterOrEqual(t, found.EndColumn, found.Column)
}

func TestCheck_ParagraphHandling(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.DisableCleanOverlapping() // avoid TL span colliding with sentence matches
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	// document-end span so it does not overlap "this this" (0..n)
	lt.AddTextLevelRuleChecker("TL", func(sents []*AnalyzedSentence) []LocalMatch {
		return []LocalMatch{{FromPos: 8, ToPos: 9, RuleID: "TL", Message: "tl"}}
	})
	// NORMAL: both
	lt.SetParagraphHandling(ParagraphNormal)
	ms := lt.Check("this this")
	ids := map[string]bool{}
	for _, m := range ms {
		ids[m.RuleID] = true
	}
	require.True(t, ids["WORD_REPEAT_RULE"])
	require.True(t, ids["TL"])

	// ONLYPARA: sentence rules skipped (checkAnalyzedSentence empty)
	lt.SetParagraphHandling(ParagraphOnlyPara)
	ms = lt.Check("this this")
	ids = map[string]bool{}
	for _, m := range ms {
		ids[m.RuleID] = true
	}
	require.False(t, ids["WORD_REPEAT_RULE"])
	require.True(t, ids["TL"])

	// ONLYNONPARA: text-level skipped
	lt.SetParagraphHandling(ParagraphOnlyNonPara)
	ms = lt.Check("this this")
	ids = map[string]bool{}
	for _, m := range ms {
		ids[m.RuleID] = true
	}
	require.True(t, ids["WORD_REPEAT_RULE"])
	require.False(t, ids["TL"])
}
