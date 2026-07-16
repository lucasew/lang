package de

import (
	"bufio"
	"embed"
	"strings"
	"sync"
)

//go:embed data/case_rule_exceptions.txt
var caseRuleExceptionsFS embed.FS

var (
	caseExcOnce sync.Once
	caseExcSet  map[string]struct{}
)

// CaseRuleExceptions returns the set of CaseRule exception phrases/regexes
// from case_rule_exceptions.txt (eigennamen_gross.txt is Premium and omitted).
func CaseRuleExceptions() map[string]struct{} {
	caseExcOnce.Do(func() {
		caseExcSet = map[string]struct{}{}
		f, err := caseRuleExceptionsFS.Open("data/case_rule_exceptions.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			caseExcSet[line] = struct{}{}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
	})
	return caseExcSet
}

// IsCaseRuleException reports whether phrase is listed as a CaseRule exception.
func IsCaseRuleException(phrase string) bool {
	_, ok := CaseRuleExceptions()[phrase]
	return ok
}
