package spelling

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Prohibited entries from GlobalSpellingTest (Java field initializers).
// Exact full-line match for expressions; per-token match after split on " ".
var (
	// prohibitedGlobalExpressions ports GlobalSpellingTest.prohibitedExpressions.
	prohibitedGlobalExpressions = []string{"Dnipro", "Dnepr"}
	// prohibitedGlobalTokens ports GlobalSpellingTest.prohibitedTokens.
	prohibitedGlobalTokens = []string{"Tolstoi", "Tolstoy", "Dostoevsky"}
)

// AvoidSomeWords ports GlobalSpellingTest.avoidSomeWords over lines already loaded
// the way Java does:
//
//	getDataBroker().getFromResourceDirAsLines(SpellingCheckRule.GLOBAL_SPELLING_FILE)
//
// Control flow mirrors Java line-for-line:
//  1. split on "#" (comment); if no parts, continue
//  2. entry = parts[0].trim()  (Java String.trim → tools.JavaStringTrim)
//  3. if entry is a prohibited expression → IllegalStateException message
//  4. tokens = entry.split(" "); if any token is prohibited → same exception style
//
// Error text matches Java IllegalStateException messages (note: message says
// "global_spelling.txt" while the resource is spelling_global.txt).
func AvoidSomeWords(lines []string) error {
	for _, line := range lines {
		parts := strings.Split(line, "#")
		// Java String.split("#") with limit 0 discards trailing empties, so a line
		// that is only "#" yields length 0. Go Split keeps empties; empty parts[0]
		// never matches the prohibited lists, so outcomes match.
		if len(parts) == 0 {
			continue
		}
		entry := tools.JavaStringTrim(parts[0])
		for _, p := range prohibitedGlobalExpressions {
			if entry == p {
				return fmt.Errorf("Do not use '%s' in global_spelling.txt. It is not a valid spelling for all languages.", entry)
			}
		}
		tokens := strings.Split(entry, " ")
		for _, token := range tokens {
			for _, p := range prohibitedGlobalTokens {
				if token == p {
					return fmt.Errorf("Do not use '%s' in global_spelling.txt. It is not a valid spelling for all languages.", token)
				}
			}
		}
	}
	return nil
}

// ValidateGlobalSpellingLines is an alias of AvoidSomeWords (historical name).
func ValidateGlobalSpellingLines(lines []string) error {
	return AvoidSomeWords(lines)
}

// LoadGlobalSpellingFileLines reads official GLOBAL_SPELLING_FILE ("spelling_global.txt")
// from the same core resource path Java DataBroker serves:
//
//	org/languagetool/resource/spelling_global.txt
//
// Discovery uses DiscoverSpellingGlobal (inspiration languagetool-core resources).
// Returns an error if the official file is not found (no invent / soft empty).
func LoadGlobalSpellingFileLines() ([]string, error) {
	path := DiscoverSpellingGlobal()
	if path == "" {
		return nil, fmt.Errorf("official %s not found (Java DataBroker resource %s under org/languagetool/resource/)", GlobalSpellingFile, GlobalSpellingFile)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	// Large phrase lines in spelling_global.txt.
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		// Java BufferedReader.readLine: strip terminators only, keep content.
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// AvoidSomeWordsOnOfficialFile ports the full GlobalSpellingTest.avoidSomeWords:
// load GLOBAL_SPELLING_FILE via resource discovery, then run the prohibited checks.
func AvoidSomeWordsOnOfficialFile() error {
	lines, err := LoadGlobalSpellingFileLines()
	if err != nil {
		return err
	}
	return AvoidSomeWords(lines)
}
