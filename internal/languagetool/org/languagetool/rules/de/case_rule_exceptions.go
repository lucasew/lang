package de

import (
	"bufio"
	"embed"
	"fmt"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/eigennamen_gross.txt data/case_rule_exceptions.txt
var caseRuleExceptionsFS embed.FS

var (
	caseExcOnce sync.Once
	caseExcSet  map[string]struct{}
)

// CaseRuleExceptions returns the set of CaseRule exception phrases/regexes.
// Load twins CaseRuleExceptions.loadExceptions paths (Java order):
//
//	/de/eigennamen_gross.txt
//	/de/case_rule_exceptions.txt
//
// no TrimSpace invent; lines that start/end with whitespace (charAt UTF-16)
// throw like Java IllegalArgumentException.
func CaseRuleExceptions() map[string]struct{} {
	caseExcOnce.Do(func() {
		caseExcSet = map[string]struct{}{}
		// Java: loadExceptions("/de/eigennamen_gross.txt", "/de/case_rule_exceptions.txt")
		for _, name := range []string{"data/eigennamen_gross.txt", "data/case_rule_exceptions.txt"} {
			loadCaseRuleExceptionFile(name)
		}
	})
	return caseExcSet
}

func loadCaseRuleExceptionFile(name string) {
	f, err := caseRuleExceptionsFS.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		// Java getFromResourceDirAsLines: strip line terminators only, not spaces.
		// Scanner drops \n; strip leftover \r from CRLF resources.
		line := strings.TrimSuffix(sc.Text(), "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Java: Character.isWhitespace(line.charAt(0)) || charAt(length-1)
		// Not Go unicode.IsSpace (NBSP U+00A0 differs).
		n := utf16LenDE(line)
		if n > 0 {
			c0 := javaCharAtDE(line, 0)
			cN := javaCharAtDE(line, n-1)
			if tools.CharacterIsWhitespace(c0) || tools.CharacterIsWhitespace(cN) {
				panic(fmt.Sprintf("Invalid line in %s, starts or ends with whitespace: '%s'", name, line))
			}
		}
		caseExcSet[line] = struct{}{}
	}
	if err := sc.Err(); err != nil {
		panic(err)
	}
}

// IsCaseRuleException reports whether phrase is listed as a CaseRule exception.
func IsCaseRuleException(phrase string) bool {
	_, ok := CaseRuleExceptions()[phrase]
	return ok
}
