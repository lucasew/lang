package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java getDescription strings for DE wrappers / embedded bases.
func TestMeta_GetDescriptions(t *testing.T) {
	require.Equal(t, "Synonyme für wiederholte Wörter.", NewGermanRepeatedWordsRule(nil).GetDescription())
	require.Equal(t, "Findet lange Sätze", NewLongSentenceRule(nil, 40).GetDescription())
	require.Equal(t, "Zusammenschreibung von Wörtern, z. B. 'CD-ROM' statt 'CD ROM'", NewGermanCompoundRule(nil).GetDescription())
	require.Equal(t, "Wiederholte Worte in aufeinanderfolgenden Sätzen", NewGermanStyleRepeatedWordRule(nil).GetDescription())
	require.Equal(t, "Mögliche Wortverwechslungen: $match", NewGermanWrongWordInContextRule(nil).GetDescription())
	require.Equal(t, "Prüft auf bestimmte falsche Wörter/Phrasen: $match", NewSimpleReplaceRule(nil).GetDescription())
	require.Equal(t, "Statistische Stilanalyse: Zu häufig genutztes Substantiv", NewStyleTooOftenUsedNounRule(nil).GetDescription())
	require.Equal(t, "Einheitliche Schreibweise für Wörter mit mehr als einer korrekten Schreibweise", NewWordCoherencyRule(nil).GetDescription())
	require.Equal(t, "Fehlendes Leerzeichen zwischen Sätzen oder nach Ordnungszahlen", NewSentenceWhitespaceRule(nil).GetDescription())
}

// Java category / ITS for DE rules that override base meta.
func TestMeta_Categories(t *testing.T) {
	// GermanWordRepeatRule: Categories.REDUNDANCY (overrides WordRepeatRule MISC).
	wr := NewGermanWordRepeatRule(nil)
	require.NotNil(t, wr.GetCategory())
	require.Equal(t, rules.NewCategoryId("REDUNDANCY"), wr.GetCategory().GetID())
	require.Equal(t, rules.ITSDuplication, wr.GetLocQualityIssueType())

	// GermanWrongWordInContextRule: CONFUSED_WORDS + "Leicht zu verwechselnde Wörter" + Misspelling.
	ww := NewGermanWrongWordInContextRule(nil)
	require.NotNil(t, ww.GetCategory())
	require.Equal(t, rules.NewCategoryId("CONFUSED_WORDS"), ww.GetCategory().GetID())
	require.Equal(t, "Leicht zu verwechselnde Wörter", ww.GetCategory().GetName())
	require.Equal(t, rules.ITSMisspelling, ww.GetLocQualityIssueType())

	// GermanSpellerRule: TYPOS + Misspelling.
	sp := NewGermanSpellerRule(nil)
	require.NotNil(t, sp.GetCategory())
	require.Equal(t, rules.NewCategoryId("TYPOS"), sp.GetCategory().GetID())
	require.Equal(t, rules.ITSMisspelling, sp.GetLocQualityIssueType())

	// de.SentenceWhitespaceRule: MISC (overrides core TYPOGRAPHY) + Whitespace.
	sw := NewSentenceWhitespaceRule(nil)
	require.NotNil(t, sw.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), sw.GetCategory().GetID())
	require.Equal(t, rules.ITSWhitespace, sw.GetLocQualityIssueType())
}
