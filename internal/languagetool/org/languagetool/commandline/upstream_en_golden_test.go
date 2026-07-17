package commandline

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// upstreamGoldenFile is produced by scripts/vendor-lt-testdata.py from official
// LanguageTool <example correction="…"> nodes. Do not hand-edit cases.
type upstreamGoldenFile struct {
	Source   string `json:"source"`
	Language string `json:"language"`
	Cases    []struct {
		Rule       string `json:"rule"`
		Text       string `json:"text"`
		Suggestion string `json:"suggestion"`
		Source     string `json:"source"`
	} `json:"cases"`
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	dir := wd
	for {
		if st, err := os.Stat(filepath.Join(dir, "testdata", "upstream")); err == nil && st.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root with testdata/upstream not found")
		}
		dir = parent
	}
}

func loadUpstreamGoldens(t *testing.T, lang string) upstreamGoldenFile {
	t.Helper()
	path := filepath.Join(repoRoot(t), "testdata", "upstream", "goldens", lang+"-examples.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err, "run scripts/vendor-lt-testdata.py --langs %s", lang)
	var doc upstreamGoldenFile
	require.NoError(t, json.Unmarshal(data, &doc))
	require.NotEmpty(t, doc.Cases, "upstream golden file empty for %s", lang)
	return doc
}

func runUpstreamGoldenSample(t *testing.T, lang string) {
	t.Helper()
	doc := loadUpstreamGoldens(t, lang)
	require.NotEmpty(t, DiscoverGrammarDir(nil), "need testdata/grammar (*-upstream-soft.xml)")

	sampleN := 40
	if os.Getenv("LANG_UPSTREAM_GOLDEN_ALL") != "" {
		sampleN = len(doc.Cases)
	}
	if sampleN > len(doc.Cases) {
		sampleN = len(doc.Cases)
	}

	type miss struct{ Rule, Text, Why string }
	var misses []miss
	passed := 0
	for i := 0; i < sampleN; i++ {
		tc := doc.Cases[i]
		var buf bytes.Buffer
		_, err := CoreGoldenHook(&buf, tc.Text, &CommandLineOptions{Language: lang})
		if err != nil {
			misses = append(misses, miss{tc.Rule, tc.Text, "hook: " + err.Error()})
			continue
		}
		var findings []Finding
		if err := json.Unmarshal(buf.Bytes(), &findings); err != nil {
			misses = append(misses, miss{tc.Rule, tc.Text, "json: " + err.Error()})
			continue
		}
		found := false
		for _, f := range findings {
			if f.Rule != tc.Rule {
				continue
			}
			found = true
			if os.Getenv("LANG_UPSTREAM_GOLDEN_STRICT") == "1" && tc.Suggestion != "" && f.Suggestion != "" {
				if strings.ToLower(tc.Suggestion) != strings.ToLower(f.Suggestion) {
					misses = append(misses, miss{tc.Rule, tc.Text, "suggestion: " + f.Suggestion + " want " + tc.Suggestion})
					found = false
				}
			}
			break
		}
		if found {
			passed++
		} else {
			misses = append(misses, miss{tc.Rule, tc.Text, "rule not in findings"})
		}
	}

	t.Logf("upstream %s goldens: passed=%d missed=%d of %d", lang, passed, len(misses), sampleN)
	if len(misses) > 0 && len(misses) <= 8 {
		for _, m := range misses {
			t.Logf("  miss %s: %q (%s)", m.Rule, m.Text, m.Why)
		}
	}
	if os.Getenv("LANG_UPSTREAM_GOLDEN_STRICT") == "1" {
		require.Empty(t, misses, "strict upstream golden misses for %s", lang)
	}
	require.Greater(t, passed, 0, "no upstream %s example matched", lang)
	minPass := sampleN * 2 / 3
	// Small golden files (few simple rules): require all but allow 1 miss.
	if sampleN < 10 {
		minPass = sampleN - 1
		if minPass < 1 {
			minPass = 1
		}
	}
	require.GreaterOrEqual(t, passed, minPass, "upstream %s pass rate too low: %d/%d", lang, passed, sampleN)
}

// TestGolden_UpstreamENExamples — vendored EN examples only (no invented fixtures).
func TestGolden_UpstreamENExamples(t *testing.T) {
	runUpstreamGoldenSample(t, "en")
}

// TestGolden_UpstreamDEExamples — vendored DE examples only.
func TestGolden_UpstreamDEExamples(t *testing.T) {
	runUpstreamGoldenSample(t, "de")
}

// TestGolden_UpstreamFRExamples — vendored FR examples only.
func TestGolden_UpstreamFRExamples(t *testing.T) {
	runUpstreamGoldenSample(t, "fr")
}
