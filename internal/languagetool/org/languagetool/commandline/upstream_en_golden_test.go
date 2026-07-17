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

// upstreamENGoldenFile is produced by scripts/vendor-lt-testdata.py from official
// LanguageTool <example correction="…"> nodes. Do not hand-edit cases.
type upstreamENGoldenFile struct {
	Source   string `json:"source"`
	Language string `json:"language"`
	Cases    []struct {
		Rule       string `json:"rule"`
		Text       string `json:"text"`
		Suggestion string `json:"suggestion"`
		Source     string `json:"source"`
	} `json:"cases"`
}

func loadUpstreamENGoldens(t *testing.T) upstreamENGoldenFile {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	dir := wd
	var path string
	for {
		cand := filepath.Join(dir, "testdata", "upstream", "goldens", "en-examples.json")
		if st, err := os.Stat(cand); err == nil && !st.IsDir() {
			path = cand
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("testdata/upstream/goldens/en-examples.json not found; run scripts/vendor-lt-testdata.py")
		}
		dir = parent
	}
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var doc upstreamENGoldenFile
	require.NoError(t, json.Unmarshal(data, &doc))
	require.NotEmpty(t, doc.Cases, "upstream golden file is empty")
	return doc
}

// TestGolden_UpstreamENExamples runs CoreGoldenHook against vendored upstream
// EN examples only — no invented fixtures.
//
// Default: sample of cases, assert majority green (soft tokenizer gaps expected).
// LANG_UPSTREAM_GOLDEN_ALL=1 — entire JSON file.
// LANG_UPSTREAM_GOLDEN_STRICT=1 — every case must match rule id.
func TestGolden_UpstreamENExamples(t *testing.T) {
	doc := loadUpstreamENGoldens(t)
	require.NotEmpty(t, DiscoverGrammarDir(nil), "need testdata/grammar (en-upstream-soft.xml)")

	sampleN := 40
	if os.Getenv("LANG_UPSTREAM_GOLDEN_ALL") != "" {
		sampleN = len(doc.Cases)
	}
	if sampleN > len(doc.Cases) {
		sampleN = len(doc.Cases)
	}

	type miss struct {
		Rule, Text, Why string
	}
	var misses []miss
	passed := 0
	for i := 0; i < sampleN; i++ {
		tc := doc.Cases[i]
		var buf bytes.Buffer
		_, err := CoreGoldenHook(&buf, tc.Text, &CommandLineOptions{Language: "en"})
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

	t.Logf("upstream EN goldens: passed=%d missed=%d of %d", passed, len(misses), sampleN)
	if len(misses) > 0 && len(misses) <= 8 {
		for _, m := range misses {
			t.Logf("  miss %s: %q (%s)", m.Rule, m.Text, m.Why)
		}
	}

	strict := os.Getenv("LANG_UPSTREAM_GOLDEN_STRICT") == "1"
	if strict {
		require.Empty(t, misses, "strict upstream golden misses")
	}
	// Soft engine is not full LT yet: require a solid majority of vendored examples.
	require.Greater(t, passed, 0, "no upstream example matched")
	minPass := sampleN * 2 / 3
	require.GreaterOrEqual(t, passed, minPass, "upstream pass rate too low: %d/%d misses=%v", passed, sampleN, misses)
}
