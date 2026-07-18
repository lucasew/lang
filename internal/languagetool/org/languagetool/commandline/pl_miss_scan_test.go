package commandline

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"testing"
)

// Debug-only: LANG_PL_MISS_SCAN=1 go test -run TestDebugPLMissScan -v
func TestDebugPLMissScan(t *testing.T) {
	if os.Getenv("LANG_PL_MISS_SCAN") == "" {
		t.Skip("set LANG_PL_MISS_SCAN=1")
	}
	doc := loadUpstreamGoldens(t, "pl")
	optionalIDs := loadOptionalUpstreamSoftRuleIDs(t, "pl")
	byRule := map[string]int{}
	passed, tried := 0, 0
	var samples []string
	for _, tc := range doc.Cases {
		if _, off := optionalIDs[tc.Rule]; off {
			continue
		}
		tried++
		var buf bytes.Buffer
		_, err := CoreGoldenHook(&buf, tc.Text, &CommandLineOptions{Language: "pl"})
		if err != nil {
			byRule[tc.Rule]++
			continue
		}
		var findings []Finding
		_ = json.Unmarshal(buf.Bytes(), &findings)
		found := false
		for _, f := range findings {
			if f.Rule == tc.Rule {
				found = true
				break
			}
		}
		if found {
			passed++
			continue
		}
		byRule[tc.Rule]++
		if len(samples) < 40 {
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
	t.Logf("pl full: passed=%d missed=%d of %d (%.1f%%)", passed, tried-passed, tried, pct)
	for i, x := range ks {
		if i >= 40 {
			break
		}
		t.Logf("miss %4d %s", x.v, x.k)
	}
	for _, s := range samples {
		t.Logf("sample %s", s)
	}
}
