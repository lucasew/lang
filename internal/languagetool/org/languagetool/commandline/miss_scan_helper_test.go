package commandline

import (
	"os"
	"sort"
	"strings"
	"testing"
)

// runDebugMissScan scores upstream soft goldens for lang, reusing one configured
// JLanguageTool (Java tests keep a single tool). Env LANG_{LANG}_MISS_SCAN=1 enables.
func runDebugMissScan(t *testing.T, lang string) {
	t.Helper()
	env := "LANG_" + strings.ToUpper(languageBaseCode(lang)) + "_MISS_SCAN"
	// EN uses LANG_EN_MISS_SCAN; regional codes use base.
	if os.Getenv(env) == "" {
		// also accept full code upper (e.g. LANG_PT_MISS_SCAN already matches base)
		t.Skip("set " + env + "=1")
	}
	runDebugMissScanMode(t, lang, false)
}

// runDebugOptionalMissScan scores only optional-upstream soft golden cases with
// SOFT_OPTIONAL enabled. Env LANG_{LANG}_OPT_MISS_SCAN=1 enables.
func runDebugOptionalMissScan(t *testing.T, lang string) {
	t.Helper()
	base := strings.ToUpper(languageBaseCode(lang))
	env := "LANG_" + base + "_OPT_MISS_SCAN"
	if os.Getenv(env) == "" {
		t.Skip("set " + env + "=1")
	}
	runDebugMissScanMode(t, lang, true)
}

func runDebugMissScanMode(t *testing.T, lang string, optionalOnly bool) {
	t.Helper()
	doc := loadUpstreamGoldens(t, lang)
	optionalIDs := loadOptionalUpstreamSoftRuleIDs(t, lang)
	opts := &CommandLineOptions{Language: lang}
	if optionalOnly {
		opts.EnabledRules = []string{"SOFT_OPTIONAL"}
	}
	lt, err := configureCoreLT(lang, opts)
	if err != nil {
		t.Fatalf("configureCoreLT: %v", err)
	}
	checker := &CoreRulesChecker{Lang: lang, lt: lt}

	byRule := map[string]int{}
	passed, tried := 0, 0
	var samples []string
	for _, tc := range doc.Cases {
		_, isOpt := optionalIDs[tc.Rule]
		if optionalOnly {
			if !isOpt {
				continue
			}
		} else if isOpt {
			continue
		}
		tried++
		ms, err := checker.Check(tc.Text)
		if err != nil {
			byRule[tc.Rule]++
			continue
		}
		found := false
		for _, m := range ms {
			if m != nil && ruleIDOfMatch(m) == tc.Rule {
				found = true
				break
			}
		}
		if found {
			passed++
			continue
		}
		byRule[tc.Rule]++
		if len(samples) < 30 {
			text := tc.Text
			if len(text) > 90 {
				text = text[:90] + "…"
			}
			samples = append(samples, tc.Rule+": "+text)
		}
	}
	type kv struct {
		k string
		v int
	}
	var ks []kv
	for k, v := range byRule {
		ks = append(ks, kv{k, v})
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i].v > ks[j].v })
	pct := 0.0
	if tried > 0 {
		pct = 100 * float64(passed) / float64(tried)
	}
	label := "full"
	if optionalOnly {
		label = "optional"
	}
	t.Logf("%s %s: passed=%d missed=%d of %d (%.1f%%)", lang, label, passed, tried-passed, tried, pct)
	for i, x := range ks {
		if i >= 30 {
			break
		}
		t.Logf("miss %4d %s", x.v, x.k)
	}
	for _, s := range samples {
		t.Logf("sample %s", s)
	}
}
