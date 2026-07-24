package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestUppercaseNounReadingFilter_FailClosedWithoutTagger(t *testing.T) {
	f := NewUppercaseNounReadingFilter()
	// soft invent removed: without TagPOS do not accept
	require.False(t, f.Accept("stand"))
}

func TestUppercaseNounReadingFilter_WithTagger(t *testing.T) {
	f := NewUppercaseNounReadingFilter()
	f.TagPOS = func(u string) []string {
		if u == "Stand" {
			return []string{"SUB:NOM:SIN:NEU"}
		}
		if u == "Laufen" {
			return []string{"VER:INF:NON"}
		}
		return nil
	}
	require.True(t, f.Accept("stand"))
	require.False(t, f.Accept("laufen"))
}

func TestUppercaseNounReadingFilter_RejectsAdjReading(t *testing.T) {
	// Java: hasPartialPosTag("SUB:") && !hasPartialPosTag("ADJ")
	f := NewUppercaseNounReadingFilter()
	f.TagPOS = func(u string) []string {
		return []string{"SUB:NOM:SIN:NEU", "ADJ:PRD:GRU:SOL"}
	}
	require.False(t, f.Accept("gut"))
}

func TestUppercaseNounReadingFilter_AcceptRuleMatch(t *testing.T) {
	f := NewUppercaseNounReadingFilter()
	f.TagPOS = func(u string) []string {
		if u == "Stand" {
			return []string{"SUB:NOM:SIN:NEU"}
		}
		return nil
	}
	m := rules.NewRuleMatch(rules.NewFakeRule("U"), nil, 0, 5, "msg")
	require.NotNil(t, f.AcceptRuleMatch(m, map[string]string{"token": "stand"}, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"token": "laufen"}, 0, nil, nil))
}

func TestUppercaseNounReadingFilter_MissingTokenPanics(t *testing.T) {
	f := NewUppercaseNounReadingFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("U"), nil, 0, 1, "msg")
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{}, 0, nil, nil)
	})
}

func TestUppercaseNounReadingFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.de.UppercaseNounReadingFilter"))
}
