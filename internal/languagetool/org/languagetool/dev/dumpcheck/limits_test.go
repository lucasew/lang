package dumpcheck

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentLimitReachedError(t *testing.T) {
	e := DocumentLimitReachedError{Limit: 10}
	require.Equal(t, 10, e.GetLimit())
	require.Contains(t, e.Error(), "10")
	require.Contains(t, e.Error(), "documents")
}

func TestErrorLimitReachedError(t *testing.T) {
	e := ErrorLimitReachedError{Limit: 5}
	require.Equal(t, 5, e.GetLimit())
	require.Contains(t, e.Error(), "errors")
}
