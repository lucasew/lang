package spelling

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of GlobalSpellingTest.avoidSomeWords
func TestGlobalSpelling_AvoidSomeWords(t *testing.T) {
	require.NoError(t, ValidateGlobalSpellingLines([]string{
		"# comment",
		"Microsoft Entra",
		"log4j",
		"car2go",
	}))
	require.Error(t, ValidateGlobalSpellingLines([]string{"Dnipro"}))
	require.Error(t, ValidateGlobalSpellingLines([]string{"Leo Tolstoy"}))
	require.Error(t, ValidateGlobalSpellingLines([]string{"Dostoevsky"}))

	cwd, _ := os.Getwd()
	dir := cwd
	for i := 0; i < 10; i++ {
		p := filepath.Join(dir, "inspiration", "languagetool", "languagetool-core", "src", "main", "resources", "org", "languagetool", "resource", "spelling_global.txt")
		f, err := os.Open(p)
		if err != nil {
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
			continue
		}
		defer f.Close()
		require.NoError(t, ValidateGlobalSpellingReader(f), "path=%s", p)
		return
	}
	t.Log("spelling_global.txt not found; synthetic checks only")
}
