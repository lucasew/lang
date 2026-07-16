package uk

import (
	"bufio"
	"embed"
	"strings"
	"sync"
)

//go:embed data/case_government.txt
var caseGovFS embed.FS

// CaseGovernmentHelper ports org.languagetool.rules.uk.CaseGovernmentHelper map loading.
type CaseGovernmentHelper struct {
	// Lemma → set of case tags (v_oru, v_zna, …)
	Map map[string]map[string]struct{}
}

var (
	caseGovOnce sync.Once
	caseGov     *CaseGovernmentHelper
)

// LoadCaseGovernmentHelper loads embedded /uk/case_government.txt once.
func LoadCaseGovernmentHelper() *CaseGovernmentHelper {
	caseGovOnce.Do(func() {
		m := map[string]map[string]struct{}{}
		f, err := caseGovFS.Open("data/case_government.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		// large lines possible
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.Split(line, " ")
			if len(parts) < 2 {
				continue
			}
			lemma := parts[0]
			vidm := strings.Split(parts[1], ":")
			set, ok := m[lemma]
			if !ok {
				set = map[string]struct{}{}
				m[lemma] = set
			}
			for _, v := range vidm {
				if v != "" {
					set[v] = struct{}{}
				}
			}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
		// static override from Java
		m["згідно з"] = map[string]struct{}{"v_oru": {}}
		caseGov = &CaseGovernmentHelper{Map: m}
	})
	return caseGov
}

// HasCaseGovernment reports whether lemma governs rvCase.
func (h *CaseGovernmentHelper) HasCaseGovernment(lemma, rvCase string) bool {
	set, ok := h.Map[lemma]
	if !ok {
		return false
	}
	_, ok = set[rvCase]
	return ok
}

// GetCaseGovernments returns case tags for a lemma.
func (h *CaseGovernmentHelper) GetCaseGovernments(lemma string) []string {
	set, ok := h.Map[lemma]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}
