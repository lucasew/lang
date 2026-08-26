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

func TestExpandLTXMLEntities_SystemEntInclude(t *testing.T) {
	dir := t.TempDir()
	entPath := filepath.Join(dir, "entities.ent")
	require.NoError(t, os.WriteFile(entPath, []byte(`<!ENTITY color "rojo">`+"\n"), 0o644))
	xml := []byte(`<?xml version="1.0"?>
<!DOCTYPE rules [
  <!ENTITY % entities SYSTEM "entities.ent">
  %entities;
]>
<rules><token>&color;</token></rules>`)
	out := string(ExpandLTXMLEntitiesWithBase(dir, xml))
	require.NotContains(t, out, "DOCTYPE")
	require.Contains(t, out, "rojo")
	require.NotContains(t, out, "&color;")
	require.NotContains(t, out, "%entities;")
}

func TestReadExpandedGrammarFile_ES_EntitiesEnt(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../.."))
	path := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/es/src/main/resources/org/languagetool/rules/es/grammar.xml")
	if _, err := os.Stat(path); err != nil {
		t.Skip("inspiration ES grammar missing")
	}
	data, err := ReadExpandedGrammarFile(path)
	require.NoError(t, err)
	s := string(data)
	// entities.ent defines shortmessage_casing etc.
	require.NotContains(t, s, "%entities;")
	require.NotContains(t, s, "DOCTYPE")
	// A known entity from entities.ent should be expanded in body
	require.NotContains(t, s, "&nbsp;") // expanded or empty, not left as unknown ref if defined
	// Load rules — should get far more than empty-entity failure
	lt := languagetool.NewJLanguageTool("es")
	n, err := RegisterGrammarFile(lt, path, "es")
	require.NoError(t, err)
	t.Logf("ES grammar rules with .ent include: %d", n)
	require.Greater(t, n, 100)
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
