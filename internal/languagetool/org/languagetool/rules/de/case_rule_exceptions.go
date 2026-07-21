package de

import (
	"bufio"
	"embed"
	"fmt"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/case_rule_exceptions.txt
var caseRuleExceptionsFS embed.FS

var (
	caseExcOnce sync.Once
	caseExcSet  map[string]struct{}
)

// CaseRuleExceptions returns the set of CaseRule exception phrases/regexes
// from case_rule_exceptions.txt (eigennamen_gross.txt is Premium and omitted).
// Load twins CaseRuleExceptions.loadExceptions: no TrimSpace invent; lines that
// start/end with whitespace (charAt UTF-16) throw like Java IllegalArgumentException.
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
					panic(fmt.Sprintf("Invalid line in case_rule_exceptions.txt, starts or ends with whitespace: '%s'", line))
				}
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
