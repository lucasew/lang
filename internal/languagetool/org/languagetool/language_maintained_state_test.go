package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.LanguageMaintainedState enum.

func TestLanguageMaintainedState(t *testing.T) {
	require.Equal(t, LanguageMaintainedState("ActivelyMaintained"), ActivelyMaintained)
	require.Equal(t, LanguageMaintainedState("LookingForNewMaintainer"), LookingForNewMaintainer)
}
