package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/SimpleReplaceVerbsRuleTest.java
// Chunk + lemma + AdjustVerbSuggestionsFilter (no surface invent).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceVerbsRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceVerbsRule(nil)
	// Inject synthesizer: return conjugated-looking form from lemma for target postag.
	// For simplicity return lemma for infinitive postags, else lemma-with-à ending for past.
	rule.Filter = NewAdjustVerbSuggestionsFilter()
	rule.Filter.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		if tok == nil || tok.GetLemma() == nil {
			return nil
		}
		lemma := *tok.GetLemma()
		// Infinitive-like postags end with N
		if len(postag) >= 3 && postag[2] == 'N' {
			return []string{lemma}
		}
		// Rough past 3s for twin permanegué / permanesqué
		switch lemma {
		case "restar":
			return []string{"restà"}
		case "estar":
			return []string{"estigué"}
		case "quedar":
			return []string{"quedà"}
		case "romandre":
			return []string{"romangué"}
		case "enfangar":
			return []string{"enfangava"}
		case "empastifar":
			return []string{"empastifava"}
		case "llepar":
			return []string{"llepava"}
		case "cagar":
			return []string{"cagava"}
		case "haver":
			return []string{"havia"}
		default:
			return []string{lemma}
		}
	}

	// permanegué — lemma permanèixer, V past
	matches := rule.Match(withIncorrectVerb("permanegué", "permanèixer", "VMIS3S00"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "restà", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "estigué", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "quedà", matches[0].GetSuggestedReplacements()[2])
	require.Equal(t, "romangué", matches[0].GetSuggestedReplacements()[3])

	matches = rule.Match(withIncorrectVerb("permanesqué", "permanèixer", "VMIS3S00"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "restà", matches[0].GetSuggestedReplacements()[0])

	// infinitive form permanéixer
	matches = rule.Match(withIncorrectVerb("permanéixer", "permanèixer", "VMN00000"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "restar", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "estar", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "quedar", matches[0].GetSuggestedReplacements()[2])
	require.Equal(t, "romandre", matches[0].GetSuggestedReplacements()[3])

	// pringava imperfect
	matches = rule.Match(withIncorrectVerb("pringava", "pringar", "VMII3S00"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "enfangava", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "empastifava", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "llepava", matches[0].GetSuggestedReplacements()[2])
	require.Equal(t, "cagava", matches[0].GetSuggestedReplacements()[3])
	// multiword "haver begut oli" — afterLemma path
	require.Contains(t, matches[0].GetSuggestedReplacements(), "havia begut oli")
}

func TestSimpleReplaceVerbsRule_FailClosedWithoutChunk(t *testing.T) {
	rule := NewSimpleReplaceVerbsRule(nil)
	// No chunk / POS: no invent surface hit on retumbar key alone.
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Va retumbar fort."))))
}

// withIncorrectVerb injects _incorrect_verb_ chunk + V.* lemma/POS (Java Morphy+chunker stand-in).
func withIncorrectVerb(surface, lemma, pos string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(surface)
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		if !strings.EqualFold(tok.GetToken(), surface) && tok.GetToken() != surface {
			// surface may equal token
			if tok.GetToken() != surface {
				// try any content token for single-word sentences
				if isCAPunct(tok.GetToken()) {
					continue
				}
			}
		}
		if isCAPunct(tok.GetToken()) {
			continue
		}
		tok.SetChunkTags([]string{incorrectVerbChunk})
		p, l := pos, lemma
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &p, &l), "test")
		break
	}
	return sent
}

func isCAPunct(s string) bool {
	return s == "." || s == "!" || s == "?" || s == ","
}
