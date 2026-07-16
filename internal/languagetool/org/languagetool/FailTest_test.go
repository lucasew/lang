package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFail_Fail(t *testing.T) {
	t.Skip("Java @Ignore: just for circleci tests")
	require.True(t, false)
}
