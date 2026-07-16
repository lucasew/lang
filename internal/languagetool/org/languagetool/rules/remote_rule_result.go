package rules

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// RemoteRuleResult ports org.languagetool.rules.RemoteRuleResult.
type RemoteRuleResult struct {
	Remote          bool
	Success         bool
	AdjustOffsets   bool
	Matches         []*RuleMatch
	Processed       map[*languagetool.AnalyzedSentence]struct{}
	sentenceMatches map[*languagetool.AnalyzedSentence][]*RuleMatch
}

// NewRemoteRuleResult groups matches by sentence.
func NewRemoteRuleResult(remote, success, adjustOffsets bool, matches []*RuleMatch, processed []*languagetool.AnalyzedSentence) *RemoteRuleResult {
	r := &RemoteRuleResult{
		Remote:          remote,
		Success:         success,
		AdjustOffsets:   adjustOffsets,
		Matches:         matches,
		Processed:       map[*languagetool.AnalyzedSentence]struct{}{},
		sentenceMatches: map[*languagetool.AnalyzedSentence][]*RuleMatch{},
	}
	for _, s := range processed {
		r.Processed[s] = struct{}{}
	}
	for _, m := range matches {
		if m == nil || m.Sentence == nil {
			continue
		}
		r.sentenceMatches[m.Sentence] = append(r.sentenceMatches[m.Sentence], m)
	}
	return r
}

func (r *RemoteRuleResult) IsRemote() bool            { return r.Remote }
func (r *RemoteRuleResult) IsSuccess() bool           { return r.Success }
func (r *RemoteRuleResult) ShouldAdjustOffsets() bool { return r.AdjustOffsets }
func (r *RemoteRuleResult) GetMatches() []*RuleMatch  { return r.Matches }

// MatchesForSentence returns matches for sentence; nil if unprocessed, empty if processed with no hits.
func (r *RemoteRuleResult) MatchesForSentence(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	if ms, ok := r.sentenceMatches[sentence]; ok {
		return ms
	}
	if _, ok := r.Processed[sentence]; ok {
		return []*RuleMatch{}
	}
	return nil
}
