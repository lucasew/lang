package uk

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSearchMatch_MAfter(t *testing.T) {
	m := NewSearchMatch("не був")
	tokens := []string{"Він", "не", "був", "тут"}
	require.Equal(t, 1, m.MAfter(tokens, 0))
	require.Equal(t, -1, m.MAfter(tokens, 2))
}

func TestSearchMatch_MBefore(t *testing.T) {
	m := NewSearchMatch("не був")
	tokens := []string{"Він", "не", "був", "тут"}
	require.Equal(t, 1, m.MBefore(tokens, 2))
}

func TestSearchMatch_ConditionLemmaPostag(t *testing.T) {
	treba := "треба"
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrLemma("треба", &treba, "adv"),
		atr("було", "verb:imperf:past:n"),
	}
	m := (&SearchMatch{IgnoreQuotes: true, Limit: -1}).
		Target(ConditionLemma(regexp.MustCompile(`^(треба|потрібно)$`)))
	require.GreaterOrEqual(t, m.MNowATR(tokens, 0), 0)

	// Java pos-1 < iCond requires SENT_START room for 2-token line at end pos 2
	m2 := NewSearchMatch("не сила")
	tokens2 := []*languagetool.AnalyzedTokenReadings{
		atr("SENT"), atr("це"), atr("не"), atr("сила"),
	}
	require.Equal(t, 2, m2.MBeforeATR(tokens2, 3))
}

func TestSearchMatch_Skip(t *testing.T) {
	// skip applies only before first target hit (Java); use Latin surfaces
	m := NewSearchMatch("a b").Skip(ConditionToken("x"))
	tokens := []string{"x", "a", "b"}
	require.Equal(t, 1, m.MAfter(tokens, 0))
}

func TestSearchMatch_TargetPostag(t *testing.T) {
	m := (&SearchMatch{IgnoreQuotes: true, Limit: 4}).
		Target(ConditionPostag(regexp.MustCompile(`.*predic.*`)))
	tokens := []*languagetool.AnalyzedTokenReadings{
		atr("було", "verb"),
		atr("видно", "adv:predic"),
		atr("супутники", "noun"),
	}
	require.Equal(t, 1, m.MAfterATR(tokens, 0))
}

func TestSearchMatch_ParenInsertFallThrough(t *testing.T) {
	// Java: on "(" with ignoreInserts, jump to ")", then evaluate condition at that token
	// (does not continue past the closing paren without a match attempt).
	m := (&SearchMatch{IgnoreQuotes: true, IgnoreInserts: true, Limit: -1}).
		Target(ConditionToken(")"))
	tokens := []*languagetool.AnalyzedTokenReadings{
		atr("("), atr("ще", "adv"), atr(")"), atr("і", "conj"),
	}
	// starting at "(" → jump to ")" → match target
	require.Equal(t, 2, m.MAfterATR(tokens, 0))
}

func TestSearchMatch_CommaInsertSkip(t *testing.T) {
	m := NewSearchMatch("a b").IgnoreInsertsOn()
	// a , зокрема , b — lemma-based insert skip (Java hasLemma зокрема|відповідно)
	atrs := []*languagetool.AnalyzedTokenReadings{
		atr("a"), atr(","), atrLemma("зокрема", strPtr("зокрема"), "adv"), atr(","), atr("b"),
	}
	require.Equal(t, 4, m.MAfterATR(atrs, 0))
}

func TestSearchMatch_MNowSetsLimitZero(t *testing.T) {
	// Java mNow = limit(0).mAfter permanently mutates Match.limit
	m := NewSearchMatch("a").WithLimit(4)
	atrs := []*languagetool.AnalyzedTokenReadings{atr("x"), atr("a")}
	require.Equal(t, 1, m.MNowATR(atrs, 0))
	require.Equal(t, 0, m.Limit)
}
