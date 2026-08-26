package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of JLanguageTool.Level enum values.

func TestLevel_Constants(t *testing.T) {
	// Exact names from JLanguageTool.Level
	require.Equal(t, Level("DEFAULT"), LevelDefault)
	require.Equal(t, Level("PICKY"), LevelPicky)
	require.Equal(t, Level("ACADEMIC"), LevelAcademic)
	require.Equal(t, Level("CLARITY"), LevelClarity)
	require.Equal(t, Level("PROFESSIONAL"), LevelProfessional)
	require.Equal(t, Level("CREATIVE"), LevelCreative)
	require.Equal(t, Level("CUSTOMER"), LevelCustomer)
	require.Equal(t, Level("JOBAPP"), LevelJobApp)
	require.Equal(t, Level("OBJECTIVE"), LevelObjective)
	require.Equal(t, Level("ELEGANT"), LevelElegant)
}
