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

// TestGolden_UpstreamOptionalDefaultOff enables SOFT_OPTIONAL so official
// default="off" style rules from *-optional-upstream-soft.xml fire.
func TestGolden_UpstreamOptionalDefaultOff(t *testing.T) {
	// ALSO_SENT_END is default=off in upstream style.xml; vendored into optional pack.
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "You will buy some eggs also.", &CommandLineOptions{
		Language:     "en",
		EnabledRules: []string{"SOFT_OPTIONAL"},
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "ALSO_SENT_END" {
			found = true
			break
		}
	}
	// Without SOFT_OPTIONAL it should stay quiet
	var buf2 bytes.Buffer
	_, err = CoreGoldenHook(&buf2, "You will buy some eggs also.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings2 []Finding
	require.NoError(t, json.Unmarshal(buf2.Bytes(), &findings2))
	for _, f := range findings2 {
		require.NotEqual(t, "ALSO_SENT_END", f.Rule, "optional rule should stay off by default")
	}
	require.True(t, found, "SOFT_OPTIONAL should enable ALSO_SENT_END: %+v", findings)
}

// TestGolden_UpstreamSimpleReplace exercises official replace.txt via soft core EN registration.
func TestGolden_UpstreamSimpleReplace(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "This is a bussiness plan.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		// SubRuleSpecificIDs may use wrong form as id, or EN_SIMPLE_REPLACE
		low := strings.ToLower(f.Rule + " " + f.Suggestion + " " + f.Message)
		if strings.Contains(low, "business") || strings.Contains(low, "bussiness") || f.Rule == "EN_SIMPLE_REPLACE" {
			found = true
			break
		}
	}
	require.True(t, found, "expected simple-replace finding for bussiness: %+v", findings)
}

// TestGolden_UpstreamCompoundRule exercises official compounds.txt via soft EN core.
func TestGolden_UpstreamCompoundRule(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "This is a case sensitive search.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_COMPOUNDS" || strings.Contains(strings.ToLower(f.Suggestion), "case-sensitive") {
			found = true
			break
		}
	}
	require.True(t, found, "expected compound finding: %+v", findings)
}

// TestGolden_UpstreamSpecificCase exercises official specific_case.txt via soft EN core.
func TestGolden_UpstreamSpecificCase(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I like harry potter books.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if strings.Contains(strings.ToLower(f.Suggestion), "harry potter") ||
			strings.Contains(strings.ToLower(f.Message), "proper noun") ||
			strings.Contains(f.Rule, "SPECIFIC_CASE") {
			found = true
			break
		}
	}
	require.True(t, found, "expected specific-case finding: %+v", findings)
}

// TestGolden_UpstreamContractionSpelling exercises official contractions.txt.
func TestGolden_UpstreamContractionSpelling(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Dont do this at home.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_CONTRACTION_SPELLING" || strings.Contains(f.Suggestion, "Don't") || strings.Contains(f.Suggestion, "don't") {
			found = true
			break
		}
	}
	require.True(t, found, "expected contraction spelling finding: %+v", findings)
}

// TestGolden_UpstreamWrongWordInContext exercises official wrongWordInContext.txt.
func TestGolden_UpstreamWrongWordInContext(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I have proscribed you a course of antibiotics.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "ENGLISH_WRONG_WORD_IN_CONTEXT" || strings.Contains(strings.ToLower(f.Suggestion), "prescribed") {
			found = true
			break
		}
	}
	require.True(t, found, "expected wrong-word-in-context finding: %+v", findings)
}

// TestGolden_UpstreamEnglishDash exercises dash compound normalization.
func TestGolden_UpstreamEnglishDash(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I wear a T – shirt daily.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if strings.Contains(f.Suggestion, "T-shirt") || strings.Contains(f.Rule, "DASH") {
			found = true
			break
		}
	}
	require.True(t, found, "expected dash rule finding: %+v", findings)
}

// TestGolden_UpstreamAmericanReplace exercises en-US British→American replace table.
func TestGolden_UpstreamAmericanReplace(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I love fish fingers for dinner.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_US_SIMPLE_REPLACE" || strings.Contains(strings.ToLower(f.Suggestion), "fish sticks") {
			found = true
			break
		}
	}
	require.True(t, found, "expected American replace finding: %+v", findings)
}

// TestGolden_UpstreamRedundancy exercises official redundancies.txt.
func TestGolden_UpstreamRedundancy(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I ate tuna fish yesterday.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_REDUNDANCY_REPLACE" || f.Suggestion == "tuna" {
			found = true
			break
		}
	}
	require.True(t, found, "expected redundancy finding: %+v", findings)
}

// TestGolden_UpstreamPlainEnglish exercises official wordiness.txt.
func TestGolden_UpstreamPlainEnglish(t *testing.T) {
	// Use a known pair from wordiness if possible — smoke that rule fires on some phrase.
	// "in the event that" is a classic plain-English target.
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Call me in the event that you need help.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_PLAIN_ENGLISH_REPLACE" || strings.Contains(strings.ToLower(f.Message+f.Suggestion), "if") {
			// may also match soft style rules; accept any plain-english id
			if f.Rule == "EN_PLAIN_ENGLISH_REPLACE" || strings.Contains(f.Rule, "PLAIN") {
				found = true
				break
			}
		}
	}
	if !found {
		// softer: rule registered and fires on added bonus path already covered by redundancy
		// try another classic: "at this point in time"
		var buf2 bytes.Buffer
		_, err = CoreGoldenHook(&buf2, "At this point in time we should leave.", &CommandLineOptions{Language: "en"})
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(buf2.Bytes(), &findings))
		for _, f := range findings {
			if f.Rule == "EN_PLAIN_ENGLISH_REPLACE" {
				found = true
				break
			}
		}
	}
	require.True(t, found, "expected plain-english finding: %+v", findings)
}

// TestGolden_UpstreamConsistentApostrophes exercises mixed ' vs ’ apostrophes.
func TestGolden_UpstreamConsistentApostrophes(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "It's a nice idea. But it doesn’t work.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_CONSISTENT_APOS" {
			found = true
			break
		}
	}
	require.True(t, found, "expected consistent apostrophe finding: %+v", findings)
}
