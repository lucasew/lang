package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SentenceWhitespaceRule ports org.languagetool.rules.de.SentenceWhitespaceRule.
type SentenceWhitespaceRule struct {
	*rules.SentenceWhitespaceRule
}

func NewSentenceWhitespaceRule(messages map[string]string) *SentenceWhitespaceRule {
	base := rules.NewSentenceWhitespaceRule(messages)
	base.RuleID = "DE_SENTENCE_WHITESPACE"
	// Java de.SentenceWhitespaceRule: super.setCategory(MISC) overrides base TYPOGRAPHY.
	base.Category = rules.CatMisc.GetCategory(messages)
	// Java: setLocQualityIssueType(Whitespace) — same as core base; keep explicit.
	base.IssueType = rules.ITSWhitespace
	base.MessageAfterSentence = "Fügen Sie zwischen Sätzen ein Leerzeichen ein."
	base.MessageAfterNumber = "Fügen Sie nach Ordnungszahlen (1., 2. usw.) ein Leerzeichen ein."
	return &SentenceWhitespaceRule{SentenceWhitespaceRule: base}
}

func (r *SentenceWhitespaceRule) GetID() string {
	if r != nil && r.SentenceWhitespaceRule != nil {
		return r.SentenceWhitespaceRule.GetID()
	}
	return "DE_SENTENCE_WHITESPACE"
}

// GetDescription ports de.SentenceWhitespaceRule.getDescription.
func (r *SentenceWhitespaceRule) GetDescription() string {
	return "Fehlendes Leerzeichen zwischen Sätzen oder nach Ordnungszahlen"
}

// GetURL ports de.SentenceWhitespaceRule constructor setUrl.
func (r *SentenceWhitespaceRule) GetURL() string {
	return "https://languagetool.org/insights/de/beitrag/grammatik-leerzeichen/#fehler-1-leerzeichen-vor-und-nach-satzzeichen"
}

func (r *SentenceWhitespaceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	// Java attaches this (DE rule) so setUrl is visible on matches.
	ms := r.SentenceWhitespaceRule.MatchList(sentences)
	for _, m := range ms {
		if m != nil {
			m.Rule = r
		}
	}
	return ms
}
