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
	doc := loadUpstreamGoldens(t, lang)
	optionalIDs := loadOptionalUpstreamSoftRuleIDs(t, lang)
	opts := &CommandLineOptions{Language: lang}
	lt, err := configureCoreLT(lang, opts)
	if err != nil {
		t.Fatalf("configureCoreLT: %v", err)
	}
	checker := &CoreRulesChecker{Lang: lang, lt: lt}

	byRule := map[string]int{}
	passed, tried := 0, 0
	var samples []string
	for _, tc := range doc.Cases {
		if _, off := optionalIDs[tc.Rule]; off {
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
	t.Logf("%s full: passed=%d missed=%d of %d (%.1f%%)", lang, passed, tried-passed, tried, pct)
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
