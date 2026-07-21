package rules

import (
	"bufio"
	"fmt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"io"
	"strconv"
	"strings"
)

// ShortDescFunc looks up an optional short description for a word (pluggable hook).
// When nil, only inline |description from the file is used.
type ShortDescFunc func(word string) *string

// ConfusionSetLoader ports org.languagetool.rules.ConfusionSetLoader.
// Loads confusion_sets.txt (UTF-8): word1[|desc]; word2[|desc]; factor
// or word1 -> word2; factor for unidirectional pairs.
type ConfusionSetLoader struct {
	WordDefs ShortDescFunc
}

func NewConfusionSetLoader(wordDefs ShortDescFunc) *ConfusionSetLoader {
	return &ConfusionSetLoader{WordDefs: wordDefs}
}

// LoadConfusionPairs parses stream into word → list of ConfusionPair.
func (l *ConfusionSetLoader) LoadConfusionPairs(r io.Reader) (map[string][]*ConfusionPair, error) {
	m := map[string][]*ConfusionPair{}
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if strings.HasPrefix(tools.JavaStringTrim(line), "#") || tools.JavaStringTrim(line) == "" {
			continue
		}
		// Java: arrow check uses replaceFirst("#.*", ""); parts use replaceFirst("\\s*#.*", "")
		codeForArrow := line
		if i := strings.Index(codeForArrow, "#"); i >= 0 {
			codeForArrow = codeForArrow[:i]
		}
		stripped := tools.JavaStringTrim(stripConfusionComment(line))
		if stripped == "" {
			continue
		}
		parts := splitConfusionLine(stripped)
		if len(parts) != 3 {
			return nil, fmt.Errorf("unexpected format line %d: %q — expected three semicolon/arrow-separated values", lineNo, line)
		}
		bidirectional := !strings.Contains(codeForArrow, " -> ")
		var confusionStrings []*ConfusionString
		loadedForSet := map[string]struct{}{}
		var prevWord string
		for _, part := range parts[:2] {
			sub := strings.SplitN(part, "|", 2)
			word := sub[0]
			if bidirectional && prevWord != "" && word < prevWord {
				return nil, fmt.Errorf("order words alphabetically per line in the confusion set file: %s: found %s after %s", line, word, prevWord)
			}
			prevWord = word
			var description *string
			if len(sub) == 2 {
				d := sub[1]
				description = &d
			}
			if _, dup := loadedForSet[word]; dup {
				return nil, fmt.Errorf("word appears twice in same confusion set: %q", word)
			}
			if description == nil && l.WordDefs != nil {
				description = l.WordDefs(word)
			}
			confusionStrings = append(confusionStrings, NewConfusionString(word, description))
			loadedForSet[word] = struct{}{}
		}
		factorStr := tools.JavaStringTrim(strings.TrimSuffix(tools.JavaStringTrim(parts[2]), ";"))
		factor, err := strconv.ParseInt(factorStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: bad factor %q: %w", lineNo, factorStr, err)
		}
		if bidirectional {
			p1 := NewConfusionPair(confusionStrings[0], confusionStrings[1], factor, false)
			addConfusionPairToMap(m, confusionStrings, p1)
			p2 := NewConfusionPair(confusionStrings[1], confusionStrings[0], factor, false)
			addConfusionPairToMap(m, confusionStrings, p2)
		} else {
			p := NewConfusionPair(confusionStrings[0], confusionStrings[1], factor, false)
			addConfusionPairToMap(m, confusionStrings, p)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

func stripConfusionComment(line string) string {
	// Java: line.replaceFirst("\\s*#.*", "")
	if i := strings.Index(line, "#"); i >= 0 {
		return strings.TrimRight(line[:i], " \t")
	}
	return line
}

func splitConfusionLine(stripped string) []string {
	// Java: split("\\s*(;|->)\\s*")
	var parts []string
	rest := stripped
	for len(parts) < 2 {
		semi := strings.Index(rest, ";")
		arrow := strings.Index(rest, "->")
		cut, sepLen := -1, 0
		if semi >= 0 && (arrow < 0 || semi < arrow) {
			cut, sepLen = semi, 1
		} else if arrow >= 0 {
			cut, sepLen = arrow, 2
		}
		if cut < 0 {
			break
		}
		parts = append(parts, tools.JavaStringTrim(rest[:cut]))
		rest = tools.JavaStringTrim(rest[cut+sepLen:])
	}
	if rest != "" {
		parts = append(parts, tools.JavaStringTrim(rest))
	}
	return parts
}

func addConfusionPairToMap(m map[string][]*ConfusionPair, confusionStrings []*ConfusionString, pair *ConfusionPair) {
	for _, cs := range confusionStrings {
		key := cs.GetString()
		m[key] = append(m[key], pair)
	}
}
