package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeBitext struct{ id string }

func (f fakeBitext) GetID() string { return f.id }

// Twin of Tools.selectBitextRules control flow (no upstream JUnit for this helper).
func TestSelectBitextRules_DisableByID(t *testing.T) {
	rules := []fakeBitext{{"A"}, {"B"}, {"C"}}
	got := SelectBitextRules(rules, []string{"B"}, nil, false)
	require.Equal(t, []fakeBitext{{"A"}, {"C"}}, got)
	// original untouched
	require.Len(t, rules, 3)
}

func TestSelectBitextRules_UseEnabledOnly_Single(t *testing.T) {
	rules := []fakeBitext{{"A"}, {"B"}, {"C"}}
	got := SelectBitextRules(rules, nil, []string{"B"}, true)
	require.Equal(t, []fakeBitext{{"B"}}, got)
}

// Java quirk: for each enabled id, every non-matching rule is queued for removal.
// With two enabled IDs every rule fails at least one equality check → empty result.
func TestSelectBitextRules_UseEnabledOnly_Multi_JavaQuirk(t *testing.T) {
	rules := []fakeBitext{{"A"}, {"B"}, {"C"}}
	got := SelectBitextRules(rules, nil, []string{"A", "B"}, true)
	require.Empty(t, got)
}

func TestSelectBitextRules_UseEnabledOnly_EmptyEnabledKeepsAll(t *testing.T) {
	rules := []fakeBitext{{"A"}, {"B"}}
	got := SelectBitextRules(rules, nil, nil, true)
	require.Equal(t, rules, got)
}

func TestSelectBitextRules_DisableUnknownIsNoop(t *testing.T) {
	rules := []fakeBitext{{"A"}}
	got := SelectBitextRules(rules, []string{"Z"}, nil, false)
	require.Equal(t, rules, got)
}
