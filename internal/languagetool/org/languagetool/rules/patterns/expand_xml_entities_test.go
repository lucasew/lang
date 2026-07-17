package patterns

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestExpandLTXMLEntities_StripsDoctypeAndExpands(t *testing.T) {
	in := []byte(`<?xml version="1.0"?>
<!DOCTYPE rules [
  <!ENTITY foo "bar">
]>
<rules><token>&foo;</token></rules>`)
	out := string(ExpandLTXMLEntities(in))
	require.NotContains(t, out, "DOCTYPE")
	require.Contains(t, out, "bar")
	require.NotContains(t, out, "&foo;")
}

func TestRegisterGrammarFile_VendoredUpstreamEN(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../.."))
	path := filepath.Join(root, "testdata/upstream/en/rules/grammar.xml")
	if _, err := os.Stat(path); err != nil {
		t.Skip("vendored EN grammar missing")
	}
	lt := languagetool.NewJLanguageTool("en")
	n, err := RegisterGrammarFile(lt, path, "en")
	require.NoError(t, err)
	t.Logf("registered %d rules from full upstream grammar.xml", n)
	require.Greater(t, n, 50)
}
