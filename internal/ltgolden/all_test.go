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
			t.Fatal("LanguageTool tree not found; init submodule")
		}
		dir = parent
	}
}

// TestPortInventory asserts we imported every LT test source (no silent drops).
func TestPortInventory(t *testing.T) {
	root := dataRoot(t)
	rules := AllRuleXMLPaths(root)
	dis := AllDisambiguationXMLPaths(root)
	java := AllJavaTestPaths(root)
	t.Logf("rule XML files: %d", len(rules))
	t.Logf("disambiguation XML files: %d", len(dis))
	t.Logf("Java *Test.java files: %d", len(java))
	if len(rules) < 50 {
		t.Fatalf("expected dozens of rule XML files, got %d", len(rules))
	}
	if len(dis) < 10 {
		t.Fatalf("expected many disambiguation.xml, got %d", len(dis))
	}
	if len(java) < 500 {
		t.Fatalf("expected hundreds of Java tests, got %d", len(java))
	}

	gCases, err := ExtractCases(rules)
	if err != nil {
		t.Fatal(err)
	}
	dCases, err := ExtractDisambigCases(dis)
	if err != nil {
		t.Fatal(err)
	}
	jCases, err := ExtractJavaCases(java)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("grammar/style examples: %d", len(gCases))
	t.Logf("disambiguation examples: %d", len(dCases))
	t.Logf("java unit strings: %d", len(jCases))
	if len(gCases) < 10000 {
		t.Fatalf("grammar examples too few: %d", len(gCases))
	}
	if len(jCases) < 1000 {
		t.Fatalf("java-derived cases too few: %d", len(jCases))
	}
}

// TestAllGroundTruth runs the full ported LT suite as a TDD scoreboard.
// Set LANG_GOLDEN_QUICK=1 for a smaller sample (still all sources, fewer cases each).
// Default runs ALL cases (may take a long time).
func TestAllGroundTruth(t *testing.T) {
	// Require explicit env (or -short) so `go test ./...` does not run ~135k cases.
	if os.Getenv("LANG_GOLDEN") != "1" && os.Getenv("LANG_GOLDEN_QUICK") != "1" && !testing.Short() {
		t.Skip("set LANG_GOLDEN=1 for full TDD, LANG_GOLDEN_QUICK=1 for a sample, or use -short")
	}
	root := dataRoot(t)
	rules := AllRuleXMLPaths(root)
	dis := AllDisambiguationXMLPaths(root)
	java := AllJavaTestPaths(root)

	gCases, err := ExtractCases(rules)
	if err != nil {
		t.Fatal(err)
	}
	dCases, err := ExtractDisambigCases(dis)
	if err != nil {
		t.Fatal(err)
	}
	jCases, err := ExtractJavaCases(java)
	if err != nil {
		t.Fatal(err)
	}

	all := make([]Case, 0, len(gCases)+len(dCases)+len(jCases))
	all = append(all, gCases...)
	all = append(all, dCases...)
	all = append(all, jCases...)
	t.Logf("TOTAL ground-truth cases: %d (grammar=%d disambig=%d java=%d)",
		len(all), len(gCases), len(dCases), len(jCases))

	quick := os.Getenv("LANG_GOLDEN_QUICK") == "1" || testing.Short()
	if quick {
		all = sampleAllKinds(all, 300)
		t.Logf("quick mode → running %d (unset LANG_GOLDEN_QUICK and omit -short for full %d)", len(all), len(gCases)+len(dCases)+len(jCases))
	}

	c, err := engine.New(root)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	results := make([]Result, 0, len(all))
	for i, tc := range all {
		if i > 0 && i%5000 == 0 {
			t.Logf("… %d/%d", i, len(all))
		}
		results = append(results, RunCase(c, tc))
	}
	sum := Summarize(results)
	elapsed := time.Since(start)
	pct := 0.0
	if sum.Total > 0 {
		pct = 100 * float64(sum.Pass) / float64(sum.Total)
	}
	t.Logf("GROUND TRUTH: %d/%d pass (%.2f%%) in %s", sum.Pass, sum.Total, pct, elapsed.Round(time.Millisecond))
	for k, v := range sum.ByKind {
		t.Logf("  %s: %d/%d", k, v[0], v[1])
	}
	t.Logf("  incorrect examples: %d/%d | correct: %d/%d",
		sum.IncorrectPass, sum.Incorrect, sum.CorrectPass, sum.Correct)

	// Write full failure report
	failPath := filepath.Join("testdata", "ground-truth-failures.txt")
	_ = os.MkdirAll("testdata", 0o755)
	// prefer repo-relative
	if wd, err := os.Getwd(); err == nil {
		// place under package dir or repo
		failPath = filepath.Join(wd, "testdata", "ground-truth-failures.txt")
	}
	f, err := os.Create(failPath)
	if err == nil {
		defer f.Close()
		fmt.Fprintf(f, "pass=%d fail=%d total=%d pct=%.2f\n\n", sum.Pass, sum.Fail, sum.Total, pct)
		n := 0
		for _, r := range results {
			if r.Pass {
				continue
			}
			n++
			if n > 5000 {
				fmt.Fprintf(f, "… truncated, %d more\n", sum.Fail-5000)
				break
			}
			fmt.Fprintf(f, "FAIL kind=%s lang=%s rule=%s src=%s\n  text=%q\n  detail=%s\n\n",
				r.Case.Kind, r.Case.Lang, r.Case.RuleID, r.Case.SourceFile, r.Case.Text, r.Detail)
		}
		t.Logf("failures: %s", failPath)
	}

	// Ground truth: report but do not hide failures. For TDD loop, fail the test if any fail
	// unless LANG_GOLDEN_ALLOW_FAIL=1 (inventory-only mode).
	// Full TDD (default for complete run): fail on any miss.
	// Quick/short inventory runs report score but do not fail the package.
	allowFail := os.Getenv("LANG_GOLDEN_ALLOW_FAIL") == "1" || testing.Short() || os.Getenv("LANG_GOLDEN_QUICK") == "1"
	if sum.Fail > 0 && !allowFail {
		t.Errorf("%d ground-truth failures (%.2f%% pass) — fix engine until 100%% (TDD ground truth)", sum.Fail, pct)
	} else if sum.Fail > 0 {
		t.Logf("NOTE: %d failures in quick/allow mode; full TDD: go test ./internal/ltgolden -run TestAllGroundTruth -count=1 -timeout 0", sum.Fail)
	}
}

func sampleAllKinds(all []Case, perKind int) []Case {
	counts := map[Kind]int{}
	var out []Case
	for _, c := range all {
		if counts[c.Kind] >= perKind {
			continue
		}
		counts[c.Kind]++
		out = append(out, c)
	}
	return out
}
