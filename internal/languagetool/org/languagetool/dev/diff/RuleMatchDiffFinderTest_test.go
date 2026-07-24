package diff

// Twin of RuleMatchDiffFinderTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func makeLM(msg, context, covered, suggestion string) *LightRuleMatch {
	return &LightRuleMatch{
		Line: 1, Column: 10, FullRuleID: "FAKE_ID1", Message: msg,
		Category: "FakeCategory", Context: context, CoveredText: covered,
		Suggestions: []string{suggestion}, RuleSource: "grammar.xml",
		Title: "mytitle", Status: StatusOn, Tags: nil, Premium: false,
	}
}

func TestRuleMatchDiffFinder_NoDiff(t *testing.T) {
	f := NewRuleMatchDiffFinder()
	l1 := []*LightRuleMatch{makeLM("my message", "context", "covered text", "suggestion")}
	l2 := []*LightRuleMatch{makeLM("my message", "context", "covered text", "suggestion")}
	require.Equal(t, "[]", DiffsString(f.GetDiffs(l1, l2)))
}

func TestRuleMatchDiffFinder_AddedMatch(t *testing.T) {
	f := NewRuleMatchDiffFinder()
	l1 := []*LightRuleMatch{}
	l2 := []*LightRuleMatch{makeLM("my message", "context", "covered text", "suggestion")}
	require.Equal(t,
		"[ADDED: oldMatch=null, newMatch=1/10 FAKE_ID1[null], msg=my message, covered=covered text, suggestions=[suggestion], title=mytitle, ctx=context]",
		DiffsString(f.GetDiffs(l1, l2)))
}

func TestRuleMatchDiffFinder_RemovedMatch(t *testing.T) {
	f := NewRuleMatchDiffFinder()
	l1 := []*LightRuleMatch{makeLM("my message", "context", "covered text", "suggestion")}
	l2 := []*LightRuleMatch{}
	require.Equal(t,
		"[REMOVED: oldMatch=1/10 FAKE_ID1[null], msg=my message, covered=covered text, suggestions=[suggestion], title=mytitle, ctx=context, newMatch=null]",
		DiffsString(f.GetDiffs(l1, l2)))
}

func TestRuleMatchDiffFinder_ModifiedMessage(t *testing.T) {
	f := NewRuleMatchDiffFinder()
	l1 := []*LightRuleMatch{makeLM("my message", "context", "covered text", "suggestion")}
	l2 := []*LightRuleMatch{makeLM("my modified message", "context", "covered text", "suggestion")}
	require.Equal(t,
		"[MODIFIED: oldMatch=1/10 FAKE_ID1[null], msg=my message, covered=covered text, suggestions=[suggestion], title=mytitle, ctx=context, newMatch=1/10 FAKE_ID1[null], msg=my modified message, covered=covered text, suggestions=[suggestion], title=mytitle, ctx=context]",
		DiffsString(f.GetDiffs(l1, l2)))
}

func TestRuleMatchDiffFinder_ModifiedSuggestions(t *testing.T) {
	f := NewRuleMatchDiffFinder()
	l1 := []*LightRuleMatch{makeLM("my message", "context", "covered text", "suggestion")}
	l2 := []*LightRuleMatch{makeLM("my message", "context", "covered text", "modified suggestion")}
	require.Equal(t,
		"[MODIFIED: oldMatch=1/10 FAKE_ID1[null], msg=my message, covered=covered text, suggestions=[suggestion], title=mytitle, ctx=context, newMatch=1/10 FAKE_ID1[null], msg=my message, covered=covered text, suggestions=[modified suggestion], title=mytitle, ctx=context]",
		DiffsString(f.GetDiffs(l1, l2)))
}

func TestRuleMatchDiffFinder_ModifiedCoveredText(t *testing.T) {
	f := NewRuleMatchDiffFinder()
	l1 := []*LightRuleMatch{makeLM("my message", "context", "covered text", "suggestion")}
	l2 := []*LightRuleMatch{makeLM("my message", "context", "modified covered text", "suggestion")}
	// different key → ADDED then REMOVED
	require.Equal(t,
		"[ADDED: oldMatch=null, newMatch=1/10 FAKE_ID1[null], msg=my message, covered=modified covered text, suggestions=[suggestion], title=mytitle, ctx=context, REMOVED: oldMatch=1/10 FAKE_ID1[null], msg=my message, covered=covered text, suggestions=[suggestion], title=mytitle, ctx=context, newMatch=null]",
		DiffsString(f.GetDiffs(l1, l2)))
}
