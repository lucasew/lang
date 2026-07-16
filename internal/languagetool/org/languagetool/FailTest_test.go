package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.FailTest (@Test @Ignore)

func TestFail_Fail(t *testing.T) {
	t.Skip("Java @Ignore: just for circleci tests")
	require.True(t, false)
}
