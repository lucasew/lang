package uk

import (
	"bufio"
	"embed"
	"regexp"
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

// UsedUInsteadOfAMsg ports CaseGovernmentHelper.USED_U_INSTEAD_OF_A_MSG.
const UsedUInsteadOfAMsg = ". Можливо, вжито невнормований родовий відмінок ч.р. з закінченням -у/-ю замість -а/-я (така тенденція є в сучасній мові)?"

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
	// Java: always merge getCustomGovs first
	for _, c := range getCustomCaseGovs(tok) {
		out[c] = struct{}{}
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
			// still check adjp:pasv on every reading (Java Pattern overload always; String path only when okStart)
			// String overload: adjp:pasv is inside the startPosTag-matched loop in Java
			continue
		}
		lemma := *token.GetLemma()
		if set, ok := h.Map[lemma]; ok {
			for c := range set {
				out[c] = struct{}{}
			}
		}
		// Java String overload: adjp:pasv adds v_oru inside matched loop
		if strings.Contains(pos, "adjp:pasv") {
			out["v_oru"] = struct{}{}
		}
	}
	return out
}

// GetCaseGovernmentsFromReadingsRE ports getCaseGovernments(readings, Pattern).
// POS filter uses Matcher.matches(); missing map lemmas on advp resolve via getAdvpVerbLemma.
func (h *CaseGovernmentHelper) GetCaseGovernmentsFromReadingsRE(tok *languagetool.AnalyzedTokenReadings, posRE *regexp.Regexp) map[string]struct{} {
	out := map[string]struct{}{}
	if h == nil || tok == nil {
		return out
	}
	for _, c := range getCustomCaseGovs(tok) {
		out[c] = struct{}{}
	}
	for _, token := range tok.GetReadings() {
		if token == nil || token.GetPOSTag() == nil {
			continue
		}
		pos := *token.GetPOSTag()
		if pos == "" {
			continue
		}
		// Java: posTag == null || matches
		if posRE != nil {
			loc := posRE.FindStringIndex(pos)
			if loc == nil || loc[0] != 0 || loc[1] != len(pos) {
				// still may add v_oru for adjp:pasv below
				if strings.Contains(pos, "adjp:pasv") {
					out["v_oru"] = struct{}{}
				}
				continue
			}
		}
		lemma := ""
		if token.GetLemma() != nil {
			lemma = *token.GetLemma()
		}
		if lemma != "" {
			if _, ok := h.Map[lemma]; !ok && strings.HasPrefix(pos, "advp") {
				lemma = advpVerbLemma(token)
			}
			if set, ok := h.Map[lemma]; ok {
				for c := range set {
					out[c] = struct{}{}
				}
			}
		}
		if strings.Contains(pos, "adjp:pasv") {
			out["v_oru"] = struct{}{}
		}
	}
	return out
}

var (
	cgMatiPosRE = regexp.MustCompile(`^verb:imperf:(?:futr|past|pres).*`)
	cgButiPosRE = regexp.MustCompile(`^verb:imperf:(?:futr|past:n|pres:s:3).*`)
	cgImpersVInfRE = regexp.MustCompile(`^verb.*(?:pres:s:3|futr:s:3|past:n).*`)
	cgNalezhytyInfRE = regexp.MustCompile(`^verb:imperf:inf.*`)
	cgBilshMenshRE = regexp.MustCompile(`^(?:по)?більшати|(?:по)?меншати$`)
	cgBilshMenshPosRE = regexp.MustCompile(`^verb.*(?:inf|pres:s:3|futr:s:3|past:n).*`)
)

// getCustomCaseGovs ports CaseGovernmentHelper.getCustomGovs (special inflection governments).
func getCustomCaseGovs(tok *languagetool.AnalyzedTokenReadings) []string {
	if tok == nil {
		return nil
	}
	var list []string
	if HasLemmaWithPosRE(tok, []string{"мати"}, cgMatiPosRE) {
		list = append(list, "v_inf")
	} else if HasLemmaWithPosRE(tok, []string{"бути"}, cgButiPosRE) {
		list = append(list, "v_inf")
	} else if HasLemmaWithPosRE(tok, []string{
		"вимагатися", "випадати", "випасти", "личити", "належати", "тягнути", "щастити",
		"плануватися", "рекомендуватися", "пропонуватися", "сподобатися", "прийтися",
		"удатися", "годитися", "доводитися",
	}, cgImpersVInfRE) {
		list = append(list, "v_inf")
	} else if HasLemmaWithPosRE(tok, []string{"належить"}, cgNalezhytyInfRE) {
		// Java list is surface "належить" as lemma list — keep twin
		list = append(list, "v_inf")
	} else if HasLemmaTokenRE(tok, cgBilshMenshRE) && HasPosTagMatches(tok, cgBilshMenshPosRE) {
		list = append(list, "v_rod")
	}
	return list
}

// advpVerbLemma ports getAdvpVerbLemma (map missing advp lemma → base verb).
func advpVerbLemma(token *languagetool.AnalyzedToken) string {
	if token == nil || token.GetLemma() == nil {
		return ""
	}
	v := *token.GetLemma()
	switch v {
	case "даючи":
		return "давати"
	case "змушуючи":
		return "змушувати"
	}
	// replaceFirst("лячи(с[яь])?", "ити$1")
	if re := regexp.MustCompile(`лячи(с[яь])?$`); re.MatchString(v) {
		return re.ReplaceAllString(v, "ити$1")
	}
	// replaceFirst("(ючи|вши)(с[яь])?", "ти$2")
	if re := regexp.MustCompile(`(ючи|вши)(с[яь])?$`); re.MatchString(v) {
		return re.ReplaceAllString(v, "ти$2")
	}
	return v
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
