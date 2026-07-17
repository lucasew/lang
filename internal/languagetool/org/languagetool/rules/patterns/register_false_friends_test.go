package patterns_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestRegisterFalseFriendsFile_Gift(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../.."))
	path := filepath.Join(root, "testdata/false-friends-soft.xml")
	lt := languagetool.NewJLanguageTool("en")
	n, err := patterns.RegisterFalseFriendsFile(lt, path, "en", "de")
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)
	m := lt.Check("This is a gift for you.")
	found := false
	for _, x := range m {
		if x.RuleID == "GIFT" {
			found = true
			require.Equal(t, "FALSEFRIENDS", x.CategoryID)
			require.NotEmpty(t, x.Suggestions)
		}
	}
	require.True(t, found, "%+v", m)
}
