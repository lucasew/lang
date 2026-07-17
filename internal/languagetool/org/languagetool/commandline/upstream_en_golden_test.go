package commandline

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
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

func listUpstreamGoldenLangs(t *testing.T) []string {
	t.Helper()
	dir := filepath.Join(repoRoot(t), "testdata", "upstream", "goldens")
	ents, err := os.ReadDir(dir)
	require.NoError(t, err)
	var langs []string
	for _, e := range ents {
		name := e.Name()
		if !strings.HasSuffix(name, "-examples.json") {
			continue
		}
		lang := strings.TrimSuffix(name, "-examples.json")
		if lang == "" {
			continue
		}
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
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
	if len(misses) > 0 && len(misses) <= 6 {
		for _, m := range misses {
			t.Logf("  miss %s: %q (%s)", m.Rule, m.Text, m.Why)
		}
	}
	if os.Getenv("LANG_UPSTREAM_GOLDEN_STRICT") == "1" {
		require.Empty(t, misses, "strict upstream golden misses for %s", lang)
	}
	require.Greater(t, passed, 0, "no upstream %s example matched", lang)
	minPass := sampleN * 2 / 3
	if sampleN < 10 {
		minPass = sampleN - 1
		if minPass < 1 {
			minPass = 1
		}
	}
	require.GreaterOrEqual(t, passed, minPass, "upstream %s pass rate too low: %d/%d", lang, passed, sampleN)
}

// TestGolden_UpstreamExamplesMatrix runs a sample of every vendored *-examples.json
// language pack (official LT examples only; no invented fixtures).
func TestGolden_UpstreamExamplesMatrix(t *testing.T) {
	langs := listUpstreamGoldenLangs(t)
	require.NotEmpty(t, langs)
	// CJK packs are vendored for data parity, but the soft Latin tokenizer
	// does not yet surface-match character-based rules. Skip until tokenizers land.
	skipSoftMatch := map[string]string{
		"ja": "Japanese needs character/token segmentation",
		"zh": "Chinese needs character/token segmentation",
	}
	// Primary langs first for faster signal; remaining still run.
	priority := map[string]int{"en": 0, "de": 1, "fr": 2, "es": 3, "pt": 4, "pl": 5, "ca": 6, "ga": 7, "ar": 8, "ro": 9}
	sort.SliceStable(langs, func(i, j int) bool {
		pi, oki := priority[langs[i]]
		pj, okj := priority[langs[j]]
		if !oki {
			pi = 100
		}
		if !okj {
			pj = 100
		}
		if pi != pj {
			return pi < pj
		}
		return langs[i] < langs[j]
	})
	// Keep CI default bounded unless full matrix requested.
	maxLangs := 12
	if os.Getenv("LANG_UPSTREAM_GOLDEN_ALL") != "" {
		maxLangs = len(langs)
	}
	if maxLangs > len(langs) {
		maxLangs = len(langs)
	}
	n := 0
	for _, lang := range langs {
		if n >= maxLangs {
			break
		}
		if reason, ok := skipSoftMatch[lang]; ok {
			t.Run(lang, func(t *testing.T) {
				t.Skip(reason)
			})
			continue
		}
		lang := lang
		n++
		t.Run(lang, func(t *testing.T) {
			runUpstreamGoldenSample(t, lang)
		})
	}
}

// Kept as aliases so existing -run filters still work.
func TestGolden_UpstreamENExamples(t *testing.T) { runUpstreamGoldenSample(t, "en") }
func TestGolden_UpstreamDEExamples(t *testing.T) { runUpstreamGoldenSample(t, "de") }
func TestGolden_UpstreamFRExamples(t *testing.T) { runUpstreamGoldenSample(t, "fr") }
