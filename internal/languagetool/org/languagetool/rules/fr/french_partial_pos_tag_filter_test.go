package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestFrenchPartialPosTagFilter_Accept(t *testing.T) {
	f := NewFrenchPartialPosTagFilter(func(partial string) []string {
		if partial == "chat" {
			return []string{"Ncms"}
		}
		if partial == "manger" {
			return []string{"Vmn"}
		}
		return nil
	})
	ok, err := f.Accept("chatons", "^(chat)ons$", "Nc.*", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = f.Accept("chatons", "^(chat)ons$", "Vm.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	// negate
	ok, err = f.Accept("chatons", "^(chat)ons$", "Vm.*", true, false, "", "")
	require.NoError(t, err)
	require.True(t, ok) // has tags, none match Vm → negate keeps
}

func TestFrenchPartialPosTagFilter_AcceptRuleMatch(t *testing.T) {
	f := NewFrenchPartialPosTagFilter(func(partial string) []string {
		if partial == "chat" {
			return []string{"Ncms"}
		}
		return nil
	})
	// build ATR for "chatons" (surface form is what the filter regex uses)
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("chatons", "Ncms", "chat", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(chat)ons$", "postag_regexp": "Nc.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.NotNil(t, out)

	// fail-closed without tagger
	f2 := NewFrenchPartialPosTagFilter(nil)
	require.Nil(t, f2.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(chat)ons$", "postag_regexp": "Nc.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestFrenchPartialPosTagFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.fr.FrenchPartialPosTagFilter"))
}
