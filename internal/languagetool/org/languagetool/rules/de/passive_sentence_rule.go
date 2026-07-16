package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PassiveSentenceRule is a surface stand-in for PassiveSentenceRule.
// Looks for forms of "werden" plus a past-participle-like token (ge…t / ge…en).
// MinPercent 0 shows all matching sentences.
type PassiveSentenceRule struct {
	Messages   map[string]string
	MinPercent int
	werden     map[string]struct{}
}

func NewPassiveSentenceRule(messages map[string]string) *PassiveSentenceRule {
	list := []string{
		"werde", "wirst", "wird", "werden", "werdet",
		"wurde", "wurdest", "wurden", "wurdet",
		"würde", "würdest", "würden", "würdet",
		"worden",
	}
	m := map[string]struct{}{}
	for _, w := range list {
		m[w] = struct{}{}
	}
	return &PassiveSentenceRule{Messages: messages, MinPercent: 0, werden: m}
}

func (r *PassiveSentenceRule) GetID() string { return "PASSIVE_SENTENCE_DE" }

func looksLikePastParticipleDE(s string) bool {
	lc := strings.ToLower(s)
	if strings.HasPrefix(lc, "ge") && (strings.HasSuffix(lc, "t") || strings.HasSuffix(lc, "en")) && len(lc) > 4 {
		return true
	}
	// separable prefix + ge (e.g. abgegeben, hergestellt)
	if strings.Contains(lc, "ge") && (strings.HasSuffix(lc, "t") || strings.HasSuffix(lc, "en")) && len(lc) > 6 {
		// weak heuristic: contains "ge" not only at start
		return true
	}
	return false
}

func (r *PassiveSentenceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r.MinPercent != 0 {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var werdenTok *languagetool.AnalyzedTokenReadings
	hasParticiple := false
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		lc := strings.ToLower(tok.GetToken())
		if _, ok := r.werden[lc]; ok && werdenTok == nil {
			werdenTok = tok
		}
		if looksLikePastParticipleDE(tok.GetToken()) {
			hasParticiple = true
		}
	}
	if werdenTok == nil || !hasParticiple {
		return nil
	}
	msg := "Passivsatz: Aktiv formulierte Sätze sprechen im Regelfall den Leser stärker an."
	rm := rules.NewRuleMatch(r, sentence, werdenTok.GetStartPos(), werdenTok.GetEndPos(), msg)
	rm.ShortMessage = "Passivsatz"
	return []*rules.RuleMatch{rm}
}
