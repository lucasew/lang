package km

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKhmerHunspellRule(t *testing.T) {
	r := NewKhmerHunspellRuleDefault()
	require.Equal(t, "HUNSPELL_RULE", r.GetID())
}
