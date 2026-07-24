package languagetool

// Twin of languagetool-core/src/test/java/org/languagetool/GlobalSpellingTest.java
//
// Java:
//   List<String> lines = getDataBroker().getFromResourceDirAsLines(SpellingCheckRule.GLOBAL_SPELLING_FILE);
//   for each line: split "#", trim parts[0], forbid expressions / tokens.
//
// The algorithmic twin lives in rules/spelling (AvoidSomeWords) because package
// spelling imports languagetool (import cycle if we call the other way). This
// file re-implements the Java test control flow against the same official
// resource path so the package path mirrors org.languagetool.GlobalSpellingTest.

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Ports SpellingCheckRule.GLOBAL_SPELLING_FILE (spelling package also defines it;
// duplicated here to avoid importing spelling).
const globalSpellingFile = "spelling_global.txt"

// Ports GlobalSpellingTest field initializers.
var (
	prohibitedExpressions = []string{"Dnipro", "Dnepr"}
	prohibitedTokens      = []string{"Tolstoi", "Tolstoy", "Dostoevsky"}
)

// Port of GlobalSpellingTest.avoidSomeWords.
func TestGlobalSpelling_AvoidSomeWords(t *testing.T) {
	path := discoverOfficialGlobalSpelling()
	require.NotEmpty(t, path,
		"official %s must be discoverable (Java DataBroker resource under org/languagetool/resource/)",
		globalSpellingFile)

	lines, err := readLines(path)
	require.NoError(t, err)
	require.NotEmpty(t, lines)

	// Same loop body as GlobalSpellingTest.avoidSomeWords.
	for _, line := range lines {
		parts := strings.Split(line, "#")
		if len(parts) == 0 {
			continue
		}
		entry := tools.JavaStringTrim(parts[0])
		for _, p := range prohibitedExpressions {
			if entry == p {
				t.Fatalf("Do not use '%s' in global_spelling.txt. It is not a valid spelling for all languages.", entry)
			}
		}
		tokens := strings.Split(entry, " ")
		for _, token := range tokens {
			for _, p := range prohibitedTokens {
				if token == p {
					t.Fatalf("Do not use '%s' in global_spelling.txt. It is not a valid spelling for all languages.", token)
				}
			}
		}
	}
}

// discoverOfficialGlobalSpelling finds the Java core resource:
// inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/spelling_global.txt
func discoverOfficialGlobalSpelling() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		p := filepath.Join(dir, "inspiration", "languagetool", "languagetool-core",
			"src", "main", "resources", "org", "languagetool", "resource", globalSpellingFile)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty resource: %s", path)
	}
	return lines, nil
}
