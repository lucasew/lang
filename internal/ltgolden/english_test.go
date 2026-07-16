package ltgolden

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lucasew/lang/internal/engine"
)

func dataRoot(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	dir := wd
	for {
		p := filepath.Join(dir, "inspiration", "languagetool")
		if st, err := os.Stat(filepath.Join(p, "languagetool-language-modules")); err == nil && st.IsDir() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("no LT data")
		}
		dir = parent
	}
}

func TestEnglishGrammarExamples(t *testing.T) {
	root := dataRoot(t)
	paths := EnglishGrammarPaths(root)
	if len(paths) == 0 {
		t.Fatal("no grammar xml")
	}
	cases, err := ExtractCases(paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) < 100 {
		t.Fatalf("too few cases: %d", len(cases))
	}
	t.Logf("extracted %d examples from %d files", len(cases), len(paths))

	// Cap for CI speed unless LANG_GOLDEN_ALL=1
	maxCases := 800
	if os.Getenv("LANG_GOLDEN_ALL") == "1" {
		maxCases = len(cases)
	}
	if len(cases) > maxCases {
		// Prefer incorrect examples first for signal
		var bad, good []Case
		for _, c := range cases {
			if c.Incorrect {
				bad = append(bad, c)
			} else {
				good = append(good, c)
			}
		}
		var subset []Case
		for _, c := range bad {
			if len(subset) >= maxCases*2/3 {
				break
			}
			subset = append(subset, c)
		}
		for _, c := range good {
			if len(subset) >= maxCases {
				break
			}
			subset = append(subset, c)
		}
		cases = subset
		t.Logf("running subset %d (set LANG_GOLDEN_ALL=1 for full suite)", len(cases))
	}

	c, err := engine.New(root)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	results := make([]Result, 0, len(cases))
	for _, tc := range cases {
		results = append(results, RunCase(c, tc))
	}
	sum := Summarize(results)
	elapsed := time.Since(start)

	pct := 0.0
	if sum.Total > 0 {
		pct = 100 * float64(sum.Pass) / float64(sum.Total)
	}
	t.Logf("GOLDEN en: %d/%d pass (%.1f%%) in %s | incorrect %d/%d | correct %d/%d",
		sum.Pass, sum.Total, pct, elapsed.Round(time.Millisecond),
		sum.IncorrectPass, sum.Incorrect, sum.CorrectPass, sum.Correct)

	// Write failures for investigation
	failPath := filepath.Join(os.TempDir(), "lang-golden-failures.txt")
	f, _ := os.Create(failPath)
	nFail := 0
	for _, r := range results {
		if r.Pass {
			continue
		}
		nFail++
		if f != nil && nFail <= 200 {
			fmt.Fprintf(f, "FAIL %s incorrect=%v marker=%v [%d,%d)\n  text: %q\n  detail: %s\n  matches: %d\n\n",
				r.Case.RuleID, r.Case.Incorrect, r.Case.HasMarker, r.Case.MarkerFrom, r.Case.MarkerTo,
				r.Case.Text, r.Detail, len(r.Matches))
		}
	}
	if f != nil {
		f.Close()
		t.Logf("first failures written to %s (%d total fails)", failPath, sum.Fail)
	}

	// Progress gate: ratchet upward over time. Start modest so the suite is useful.
	// Full 1:1 requires ~100%. Track incorrect-example recall especially.
	minPassPct := 70.0
	if v := os.Getenv("LANG_GOLDEN_MIN_PCT"); v != "" {
		fmt.Sscanf(v, "%f", &minPassPct)
	}
	if pct < minPassPct {
		t.Fatalf("pass rate %.1f%% below gate %.1f%% — keep climbing toward 1:1", pct, minPassPct)
	}
}

func TestCoreJavaRuleExamples(t *testing.T) {
	// Port of MultipleWhitespaceRuleTest + WordRepeat style cases
	root := dataRoot(t)
	c, err := engine.New(root)
	if err != nil {
		t.Fatal(err)
	}
	type tc struct {
		text    string
		rule    string
		wantHit bool
	}
	cases := []tc{
		{"This is a test sentence.", "WHITESPACE_RULE", false},
		{"This  is a test sentence.", "WHITESPACE_RULE", true},
		{"This is a test   sentence.", "WHITESPACE_RULE", true},
		{"Multiple tabs\t\tare okay", "WHITESPACE_RULE", false},
		{"the the cat", "WORD_REPEAT_RULE", true},
		{"the cat", "WORD_REPEAT_RULE", false},
		{"The unicode standard", "UNICODE_CASING", true},
		{"The Unicode standard", "UNICODE_CASING", false},
	}
	for _, tc := range cases {
		out, err := c.Check("t", tc.text, engine.Options{Language: "en-US"})
		if err != nil {
			t.Fatal(err)
		}
		hit := false
		for _, f := range out.Findings {
			if f.Rule == tc.rule || (len(f.Rule) > len(tc.rule) && f.Rule[:len(tc.rule)] == tc.rule) {
				hit = true
				break
			}
		}
		if hit != tc.wantHit {
			t.Errorf("%q rule %s: hit=%v want %v findings=%v", tc.text, tc.rule, hit, tc.wantHit, out.Findings)
		}
	}
}
