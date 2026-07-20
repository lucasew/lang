package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadLtBuildInfo(t *testing.T) {
	info := LoadLtBuildInfo("OS", map[string]string{
		"git.build.time":       "2024-01-01T12:00:00+0000",
		"git.commit.id.abbrev": "abc1234",
		"git.build.version":    "6.0",
	})
	require.Equal(t, "abc1234", *info.GetShortGitId())
	require.Equal(t, "6.0", *info.GetVersion())
	require.Equal(t, "2024-01-01 12:00:00 +0000", *info.GetBuildDate())
	empty := LoadLtBuildInfo("PREMIUM", nil)
	require.Nil(t, empty.GetVersion())
	require.Nil(t, empty.GetBuildDate())
}

func TestLoadLtBuildInfoZulu(t *testing.T) {
	info := LoadLtBuildInfo("OS", map[string]string{
		"git.build.time": "2024-01-01T12:00:00Z",
	})
	require.Equal(t, "2024-01-01 12:00:00 +0000", *info.GetBuildDate())
}
