package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// MissingVerbRule ports org.languagetool.rules.de.MissingVerbRule.
// Requires VER POS tags and sentence-start re-tagging; without a German tagger
// Match is a no-op (avoids false positives on untagged tokens).
type MissingVerbRule struct {
	Messages map[string]string
}

func NewMissingVerbRule(messages map[string]string) *MissingVerbRule {
	return &MissingVerbRule{Messages: messages}
}

func (r *MissingVerbRule) GetID() string { return "MISSING_VERB" }

func (r *MissingVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	// Full fidelity needs VER:* tags (see Java MissingVerbRule).
	return nil
}
