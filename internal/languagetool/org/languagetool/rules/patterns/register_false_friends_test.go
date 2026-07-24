package patterns_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func repoRootFromHere() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../.."))
}

// Twin: load official false-friends.xml (Java classpath rules/false-friends.xml).
func TestRegisterFalseFriendsFile_Official(t *testing.T) {
	root := repoRootFromHere()
	// Prefer inspiration (Java resource) over incomplete testdata extracts.
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/rules/false-friends.xml"),
		filepath.Join(root, "testdata/upstream/false-friends.xml"),
		filepath.Join(root, "testdata/upstream/false-friends-nodtd.xml"),
	}
	var lastErr error
	for _, path := range candidates {
		if st, err := os.Stat(path); err != nil || !st.Mode().IsRegular() {
			continue
		}
		// Soft invent files must never be the preferred path.
		require.NotContains(t, path, "soft.xml")
		lt := languagetool.NewJLanguageTool("en")
		n, err := patterns.RegisterFalseFriendsFile(lt, path, "en", "de")
		if err != nil {
			lastErr = err
			t.Logf("%s: err=%v", path, err)
			continue
		}
		t.Logf("%s: registered %d false-friend rules", path, n)
		require.Greater(t, n, 10, "expected many en/de pairs from official FF")
		// Prefer first successful load (inspiration first)
		if filepath.Base(filepath.Dir(path)) == "rules" || n > 10 {
			return
		}
	}
	if lastErr != nil {
		t.Fatalf("official false-friends not loadable: %v", lastErr)
	}
	t.Fatal("no official false-friends file found")
}

func TestRegisterFalseFriendsFile_GiftPair(t *testing.T) {
	root := repoRootFromHere()
	path := filepath.Join(root, "inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/rules/false-friends.xml")
	if st, err := os.Stat(path); err != nil || !st.Mode().IsRegular() {
		t.Skip("inspiration false-friends.xml missing")
	}
	lt := languagetool.NewJLanguageTool("en")
	n, err := patterns.RegisterFalseFriendsFile(lt, path, "en", "de")
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)
}
