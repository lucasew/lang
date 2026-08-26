package wikipedia

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyMatchesToMarkup_Identity(t *testing.T) {
	qc := NewWikipediaQuickCheck()
	markup := "Die CD ROM."
	wiki := NewMediaWikiContent(markup, "2012-11-11T20:00:00")
	// identity plain
	from := 4 // "CD ROM"
	to := 10
	res := qc.ApplyMatchesToMarkup(wiki, markup, []MatchSpan{{
		FromPos:               from,
		ToPos:                 to,
		SuggestedReplacements: []string{"CD-ROM"},
	}}, NewErrorMarker("<err>", "</err>"))
	require.Equal(t, "2012-11-11T20:00:00", res.GetLastEditTimestamp())
	require.Equal(t, 0, res.GetInternalErrorCount())
	require.Len(t, res.GetAppliedRuleMatches(), 1)
	apps := res.GetAppliedRuleMatches()[0].GetRuleMatchApplications()
	require.NotEmpty(t, apps)
	require.Contains(t, apps[0].GetTextWithCorrection(), "CD-ROM")
	require.Contains(t, apps[0].GetTextWithCorrection(), "<err>")
}

func TestMarkupAwareWikipediaResult_InternalError(t *testing.T) {
	qc := NewWikipediaQuickCheck()
	wiki := NewMediaWikiContent("abc", "t")
	// out-of-range match → internal error
	res := qc.ApplyMatchesToMarkup(wiki, "abc", []MatchSpan{{
		FromPos:               10,
		ToPos:                 20,
		SuggestedReplacements: []string{"x"},
	}}, NewErrorMarker("<e>", "</e>"))
	require.Equal(t, 1, res.GetInternalErrorCount())
	require.Empty(t, res.GetAppliedRuleMatches())
}
