package uk

import (
	"bufio"
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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
		// Java CaseGovernmentHelper static: merge DERIVATIVES_MAP into CASE_GOVERNMENT_MAP
		// (derivative key inherits government of each base verb).
		for deriv, verbs := range loadDerivats() {
			set := map[string]struct{}{}
			if existing, ok := m[deriv]; ok {
				for c := range existing {
					set[c] = struct{}{}
				}
			}
			for v := range verbs {
				if rvs, ok := m[v]; ok {
					for c := range rvs {
						set[c] = struct{}{}
					}
				}
			}
			// Java may leave empty sets; skip empty so lookup stays useful (no invent cases).
			if len(set) > 0 {
				m[deriv] = set
			}
		}
		caseGov = &CaseGovernmentHelper{Map: m}
	})
	return caseGov
}

// GetCaseGovernmentsFromReadings ports CaseGovernmentHelper.getCaseGovernments(readings, startPosTag).
// startPosTag is a POS prefix (e.g. "adv", "prep", "verb"); "verb" with advp first reading → "advp".
func (h *CaseGovernmentHelper) GetCaseGovernmentsFromReadings(tok *languagetool.AnalyzedTokenReadings, startPosTag string) map[string]struct{} {
	out := map[string]struct{}{}
	if h == nil || tok == nil || startPosTag == "" {
		return out
	}
	rds := tok.GetReadings()
	if startPosTag == "verb" && len(rds) > 0 && rds[0] != nil && rds[0].GetPOSTag() != nil {
		if strings.HasPrefix(*rds[0].GetPOSTag(), "advp") {
			startPosTag = "advp"
		}
	}
	for _, token := range rds {
		if token == nil || token.GetLemma() == nil || token.GetPOSTag() == nil {
			continue
		}
		pos := *token.GetPOSTag()
		// Java: hasNoTag skip — Go has no hasNoTag; empty POS treated as none
		if pos == "" {
			continue
		}
		okStart := strings.HasPrefix(pos, startPosTag)
		if startPosTag == "prep" && pos == "<prep>" {
			okStart = true
		}
		if !okStart {
			continue
		}
		lemma := *token.GetLemma()
		if set, ok := h.Map[lemma]; ok {
			for c := range set {
				out[c] = struct{}{}
			}
		}
		// Java: adjp:pasv adds v_oru
		if strings.Contains(pos, "adjp:pasv") {
			out["v_oru"] = struct{}{}
		}
	}
	return out
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
