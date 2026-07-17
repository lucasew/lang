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

func TestRegisterFalseFriendsFile_Gift(t *testing.T) {
	path := filepath.Join(repoRootFromHere(), "testdata/false-friends-soft.xml")
	lt := languagetool.NewJLanguageTool("en")
	n, err := patterns.RegisterFalseFriendsFile(lt, path, "en", "de")
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)
}

func TestRegisterFalseFriendsFile_UpstreamVendored(t *testing.T) {
	root := repoRootFromHere()
	// Prefer DOCTYPE-stripped copy generated for soft loaders.
	candidates := []string{
		filepath.Join(root, "testdata/upstream/false-friends-nodtd.xml"),
		filepath.Join(root, "testdata/upstream/false-friends.xml"),
	}
	var lastErr error
	for _, path := range candidates {
		if st, err := os.Stat(path); err != nil || !st.Mode().IsRegular() {
			continue
		}
		lt := languagetool.NewJLanguageTool("en")
		n, err := patterns.RegisterFalseFriendsFile(lt, path, "en", "de")
		if err != nil {
			lastErr = err
			t.Logf("%s: err=%v", path, err)
			continue
		}
		t.Logf("%s: registered %d false-friend rules", path, n)
		require.Greater(t, n, 10, "expected many en/de pairs from upstream FF")
		return
	}
	if lastErr != nil {
		t.Fatalf("upstream false-friends not loadable: %v", lastErr)
	}
	t.Fatal("no upstream false-friends file found; run scripts/vendor-lt-testdata.py")
}
