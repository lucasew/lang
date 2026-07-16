package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadLtBuildInfo(t *testing.T) {
	info := LoadLtBuildInfo("OS", map[string]string{
		"git.build.time":       "2024-01-01T12:00:00Z",
		"git.commit.id.abbrev": "abc1234",
		"git.build.version":    "6.0",
	})
	require.Equal(t, "abc1234", *info.GetShortGitId())
	require.Equal(t, "6.0", *info.GetVersion())
	empty := LoadLtBuildInfo("PREMIUM", nil)
	require.Nil(t, empty.GetVersion())
}
