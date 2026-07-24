package rules

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// TestRemoteRule ports org.languagetool.rules.TestRemoteRule for integration tests.
// Produces a trivial match per sentence after optional waitTime (ms) from config options.
type TestRemoteRule struct {
	Config   *RemoteRuleConfig
	WaitTime time.Duration
}

func NewTestRemoteRule(config *RemoteRuleConfig) *TestRemoteRule {
	wait := time.Millisecond
	if config != nil && config.Options != nil {
		if v, ok := config.Options["waitTime"]; ok {
			if n, err := time.ParseDuration(v + "ms"); err == nil {
				wait = n
			} else if n, err := time.ParseDuration(v); err == nil {
				wait = n
			}
		}
	}
	return &TestRemoteRule{Config: config, WaitTime: wait}
}

func (r *TestRemoteRule) GetID() string {
	if r.Config != nil && r.Config.RuleID != "" {
		return r.Config.RuleID
	}
	return "TEST_REMOTE_RULE"
}

func (r *TestRemoteRule) GetDescription() string { return "TEST REMOTE RULE" }

// Execute produces RemoteRuleResult with one match (0–1) per sentence.
func (r *TestRemoteRule) Execute(sentences []*languagetool.AnalyzedSentence) *RemoteRuleResult {
	if r.WaitTime > 0 {
		time.Sleep(r.WaitTime)
	}
	var matches []*RuleMatch
	for _, s := range sentences {
		matches = append(matches, NewRuleMatch(r, s, 0, 1, "Test match"))
	}
	return NewRemoteRuleResult(true, true, true, matches, sentences)
}

// Fallback returns empty remote failure result.
func (r *TestRemoteRule) Fallback(sentences []*languagetool.AnalyzedSentence) *RemoteRuleResult {
	return NewRemoteRuleResult(false, false, false, nil, sentences)
}
