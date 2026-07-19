package de

// Twin of SimilarNameRuleTest (surface name heuristic).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimilarNameRule_Rule(t *testing.T) {
	rule := NewSimilarNameRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	require.Equal(t, 1, matchN("Hier steht Angela Müller. Im nächsten Satz dann Miller."))
	require.Equal(t, 0, matchN("Hier steht Angela Müller. Im nächsten Satz dann Müllers Ehemann."))
	require.Equal(t, 0, matchN("Hier steht Angela Müller. Dann Mulla, nicht ähnlich genug."))
	require.Equal(t, 0, matchN("Ein Mikrocontroller, bei Mikrocontrollern"))
	require.Equal(t, 0, matchN("Hier steht das Rad Deiner Freundin. Und Deinem Hund geht es gut?"))
	require.Equal(t, -1, rule.MinToCheckParagraph())

	// Java: isPosTagUnknown only (not invent !isTagged for multi-reading untagged).
	// Tagged non-EIG SUB must not be treated as a name.
	sub := atrWithPOS("Miller", "SUB:NOM:SIN:MAS", "Miller")
	require.False(t, isMaybeName(sub))
	// Single null-POS reading → isPosTagUnknown (AnalyzePlain tokens).
	unk := languagetool.AnalyzePlain("Miller").GetTokensWithoutWhitespace()
	// find Miller token
	found := false
	for _, tok := range unk {
		if tok != nil && tok.GetToken() == "Miller" {
			require.True(t, tok.IsPosTagUnknown())
			require.True(t, isMaybeName(tok))
			found = true
		}
	}
	require.True(t, found)
}
