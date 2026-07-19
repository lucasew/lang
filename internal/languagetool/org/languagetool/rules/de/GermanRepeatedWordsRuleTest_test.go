package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestGermanRepeatedWordsRule(t *testing.T) {
	rule := NewGermanRepeatedWordsRule(nil)
	// Java match uses AnalyzedToken lemmas (not surface invent). Morph tokens with lemma außerdem.
	ss := languagetool.SentenceStartTagName
	s1 := languagetool.NewAnalyzedSentence(withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Außerdem", "ADV", "außerdem"),
		atrWithPOS("regnet", "VER:3:SIN:PRÄ:SFT", "regnen"),
		atrWithPOS("es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Außerdem", "ADV", "außerdem"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("kalt", "ADJ:PRD:GRU", "kalt"),
		atrWithPOS(".", "PKT", "."),
	))
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.Equal(t, 1, len(ms))
	// Java: ruleId + "_" + StringTools.toId("außerdem", de) → ß→SS → AUSSERDEM
	require.Equal(t, "DE_REPEATEDWORDS_AUSSERDEM", ms[0].GetSpecificRuleId())
	// Java getShortMessage / RuleMatch constructor shortMessage
	require.Equal(t, "Stil: Wortwiederholung", ms[0].GetShortMessage())
	require.Equal(t, "Synonyme für wiederholte Wörter.", rule.GetDescription())
	require.Equal(t, 1, rule.MinToCheckParagraph())
	// Java JSON id = getSpecificRuleId(); ToLocalMatches must not drop lemma suffix
	lm := rules.ToLocalMatches(ms)
	require.Len(t, lm, 1)
	require.Equal(t, "DE_REPEATEDWORDS_AUSSERDEM", lm[0].RuleID)
	require.Equal(t, "Stil: Wortwiederholung", lm[0].ShortMessage)
}

// Java isException: mid-sentence capitalized form is ignored (German nouns).
func TestGermanRepeatedWordsRule_IsExceptionCapitalizedMidSentence(t *testing.T) {
	// "Leider" at sentence start counted; mid-sentence "Leider" would be isCapitalized && !sentStart.
	// Two sentence-start "leider" lowercased mid-text won't appear; use morph tokens.
	ss := languagetool.SentenceStartTagName
	// Sent1: Leider (start) … .
	// Sent2: Es war leider …  — lowercase leider mid-sentence can match.
	// Cap mid: Es war Leider — should be exception (not match).
	s1 := languagetool.NewAnalyzedSentence(withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Leider", "ADV", "leider"),
		atrWithPOS("regnet", "VER:3:SIN:PRÄ:SFT", "regnen"),
		atrWithPOS("es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS(".", "PKT", "."),
	))
	s2cap := languagetool.NewAnalyzedSentence(withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("Leider", "ADV", "leider"), // mid-sentence capitalized
		atrWithPOS("kalt", "ADJ:PRD:GRU", "kalt"),
		atrWithPOS(".", "PKT", "."),
	))
	rule := NewGermanRepeatedWordsRule(nil)
	// isCapitalized(Leider) mid-sentence → exception → no match (lemma path still uses lemma "leider")
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2cap})))

	s2low := languagetool.NewAnalyzedSentence(withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("leider", "ADV", "leider"),
		atrWithPOS("kalt", "ADJ:PRD:GRU", "kalt"),
		atrWithPOS(".", "PKT", "."),
	))
	// First sentence "Leider" at start is kept; second "leider" not exception → match
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2low})))
}

func TestGermanRepeatedWordsIsException_EIG(t *testing.T) {
	eig := atrWithPOS("Berlin", "EIG:NOM:SIN:NEU", "Berlin")
	toks := []*languagetool.AnalyzedTokenReadings{eig}
	require.True(t, germanRepeatedWordsIsException(toks, 0, true, true, false))
}

func TestGermanRepeatedWords_CategoryAndIssueType(t *testing.T) {
	r := NewGermanRepeatedWordsRule(nil)
	require.NotNil(t, r.GetCategory())
	require.Equal(t, rules.NewCategoryId("REPETITIONS_STYLE"), r.GetCategory().GetID())
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())
	require.False(t, r.HasTag(rules.TagPicky), "DE has no picky tag in Java")
}
