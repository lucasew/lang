package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAbstractPatternRule_ShortMessageIsLongerThanErrorMessage(t *testing.T) {
	// Detection helper: short message longer than message is a config smell.
	r := NewAbstractPatternRule("ID", "desc", "en", nil, false)
	r.Message = "short"
	r.ShortMessage = "this short message is actually longer than message"
	require.Greater(t, len(r.ShortMessage), len(r.Message))
}
