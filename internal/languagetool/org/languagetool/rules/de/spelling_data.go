package de

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SpellingData ports org.languagetool.rules.de.SpellingData (package-private in Java).
// Old→new spelling maps from CSV, built into Aho-Corasick tries (body + sentence-start).
// ß→ss lemma expansion uses GermanSynthesizer.INSTANCE.synthesizeForPosTags (same as Java).
type SpellingData struct {
	trie              *SpellingAhoTrie
	sentenceStartTrie *SpellingAhoTrie
}

// SpellingAhoHit ports AhoCorasickDoubleArrayTrie.Hit<String> (begin/end are UTF-16 offsets).
type SpellingAhoHit struct {
	Begin, End int
	Value      string
}

// SpellingAhoTrie ports AhoCorasickDoubleArrayTrie<String> for SpellingData lookups.
// ParseText match semantics match multi-pattern exact search (Java hankcs DAT).
type SpellingAhoTrie struct {
	root *spellingACNode
	// dict is the LinkedHashMap twin used to build the trie (old → new).
	dict map[string]string
}

type spellingACNode struct {
	next map[uint16]*spellingACNode
	fail *spellingACNode
	// out: patterns ending here (UTF-16 length + value); includes fail-link outputs.
	out []spellingACOut
}

type spellingACOut struct {
	lenU16 int
	value  string
}

// NewSpellingData ports SpellingData(String filePath).
// Loads lines via the same resource path as Java:
// JLanguageTool.getDataBroker().getFromResourceDirAsLines(filePath)
// (OldSpellingRule uses "/de/alt_neu.csv" under resource/de).
// Panics on load/sanity failure (Java constructor throws RuntimeException).
func NewSpellingData(filePath string) *SpellingData {
	data, err := readSpellingDataResource(filePath)
	if err != nil {
		panic(err)
	}
	lines := splitResourceLines(string(data))
	expand := spellingDataExpandForms()
	body, err := getCoherencyMap(lines, filePath, false, expand)
	if err != nil {
		panic(err)
	}
	sent, err := getCoherencyMap(lines, filePath, true, expand)
	if err != nil {
		panic(err)
	}
	return &SpellingData{
		trie:              buildSpellingAhoTrie(body),
		sentenceStartTrie: buildSpellingAhoTrie(sent),
	}
}

// newSpellingDataFromLines builds SpellingData from pre-split CSV lines (tests / internal).
// expand mirrors GermanSynthesizer.INSTANCE.synthesizeForPosTags; nil skips expansion
// (only for unit tests without ß→ss pairs — production NewSpellingData always wires INSTANCE).
func newSpellingDataFromLines(lines []string, filePath string, expand func(string) []string) (*SpellingData, error) {
	if filePath == "" {
		filePath = "spelling data"
	}
	body, err := getCoherencyMap(lines, filePath, false, expand)
	if err != nil {
		return nil, err
	}
	sent, err := getCoherencyMap(lines, filePath, true, expand)
	if err != nil {
		return nil, err
	}
	return &SpellingData{
		trie:              buildSpellingAhoTrie(body),
		sentenceStartTrie: buildSpellingAhoTrie(sent),
	}, nil
}

// GetTrie ports SpellingData.getTrie().
func (d *SpellingData) GetTrie() *SpellingAhoTrie {
	if d == nil {
		return nil
	}
	return d.trie
}

// GetSentenceStartTrie ports SpellingData.getSentenceStartTrie().
func (d *SpellingData) GetSentenceStartTrie() *SpellingAhoTrie {
	if d == nil {
		return nil
	}
	return d.sentenceStartTrie
}

// Dict returns the underlying old→new map used to build this trie.
func (t *SpellingAhoTrie) Dict() map[string]string {
	if t == nil {
		return nil
	}
	return t.dict
}

// ParseText ports AhoCorasickDoubleArrayTrie.parseText.
// Hit.Begin/End are UTF-16 code-unit offsets into text (Java String indices).
func (t *SpellingAhoTrie) ParseText(text string) []SpellingAhoHit {
	if t == nil || t.root == nil || text == "" {
		return nil
	}
	u := utf16.Encode([]rune(text))
	var hits []SpellingAhoHit
	n := t.root
	for i := 0; i < len(u); i++ {
		b := u[i]
		for n != t.root && n.next[b] == nil {
			n = n.fail
			if n == nil {
				n = t.root
				break
			}
		}
		if ch := n.next[b]; ch != nil {
			n = ch
		} else {
			n = t.root
		}
		for _, o := range n.out {
			begin := i + 1 - o.lenU16
			if begin < 0 {
				continue
			}
			hits = append(hits, SpellingAhoHit{
				Begin: begin,
				End:   i + 1,
				Value: o.value,
			})
		}
	}
	return hits
}

// Lookup is a convenience for exact map lookup on the body trie (tests).
func (d *SpellingData) Lookup(old string) (string, bool) {
	if d == nil || d.trie == nil || d.trie.dict == nil {
		return "", false
	}
	v, ok := d.trie.dict[old]
	return v, ok
}

// Map returns body-mode coherency map (old → new). Kept for twin tests and OldSpelling.
func (d *SpellingData) Map() map[string]string {
	if d == nil || d.trie == nil {
		return nil
	}
	return d.trie.dict
}

// SentenceStartMap returns sentence-start-mode coherency map.
func (d *SpellingData) SentenceStartMap() map[string]string {
	if d == nil || d.sentenceStartTrie == nil {
		return nil
	}
	return d.sentenceStartTrie.dict
}

// readSpellingDataResource loads filePath as under resource dir.
// Java: getDataBroker().getFromResourceDirAsLines("/de/alt_neu.csv")
// → org/languagetool/resource/de/alt_neu.csv (Official LT resource tree or embed).
func readSpellingDataResource(filePath string) ([]byte, error) {
	// OldSpellingRule.FILE_PATH and any "/de/…" resource-dir path.
	base := strings.TrimPrefix(filePath, "/")
	if i := strings.LastIndex(base, "/"); i >= 0 {
		// "/de/alt_neu.csv" → "alt_neu.csv" under resource/de
		base = base[i+1:]
	}
	if base == "alt_neu.csv" {
		return readAltNeuCSV()
	}
	if root := DiscoverGermanResourceDir(); root != "" && base != "" {
		if b, err := os.ReadFile(filepath.Join(root, base)); err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("SpellingData resource not found: %s", filePath)
}

// splitResourceLines ports BufferedReader.readLine loop (no content trim).
func splitResourceLines(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	if content == "" {
		return nil
	}
	// BufferedReader does not yield an empty line for a final trailing newline alone.
	if strings.HasSuffix(content, "\n") {
		content = content[:len(content)-1]
	}
	return strings.Split(content, "\n")
}

// spellingDataExpandForms ports GermanSynthesizer.INSTANCE.synthesizeForPosTags(lemma, s -> true).
// Returns nil only when the official german_synth.dict is not discoverable (Java always has it).
func spellingDataExpandForms() func(string) []string {
	if gs := openDiscoveredGermanSynthesizer(); gs != nil {
		return func(oldSpelling string) []string {
			return gs.SynthesizeForPosTags(oldSpelling, func(string) bool { return true })
		}
	}
	if base := openDiscoveredGermanSynthBase(); base != nil {
		return func(oldSpelling string) []string {
			return base.SynthesizeForPosTags(oldSpelling, func(string) bool { return true })
		}
	}
	return nil
}

// getCoherencyMap ports SpellingData.getCoherencyMap(filePath, sentStartMode).
func getCoherencyMap(lines []string, filePath string, sentStartMode bool, expand func(string) []string) (map[string]string, error) {
	// Java LinkedHashMap — Go map is fine for key/value fidelity; AC match does not need insert order.
	coherencyMap := make(map[string]string)
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) < 2 {
			return nil, fmt.Errorf("unexpected format in file %s: %s", filePath, line)
		}
		oldSpelling := parts[0]
		newSpelling := parts[1]
		if err := sanityChecks(filePath, line, oldSpelling, newSpelling, coherencyMap); err != nil {
			return nil, err
		}
		if sentStartMode && startsWithLowercase(oldSpelling) && startsWithLowercase(newSpelling) {
			// lowercase words can be uppercase at sentence start:
			coherencyMap[tools.UppercaseFirstChar(oldSpelling)] = tools.UppercaseFirstChar(newSpelling)
		} else {
			coherencyMap[oldSpelling] = newSpelling
		}
		// ß→ss synthesizer expansion (Java SpellingData L68–79)
		if strings.Contains(oldSpelling, "ß") && strings.ReplaceAll(oldSpelling, "ß", "ss") == newSpelling {
			if expand != nil {
				for _, form := range expand(oldSpelling) {
					if !strings.Contains(form, "ss") { // avoid e.g. "Schlüsse" as form of "Schluß"
						coherencyMap[form] = strings.ReplaceAll(form, "ß", "ss")
					}
				}
			}
		}
	}
	return coherencyMap, nil
}

func startsWithLowercase(s string) bool {
	// Java SpellingData: StringTools.startsWithLowercase (UTF-16 charAt(0))
	return tools.StartsWithLowercase(s)
}

func sanityChecks(filePath, line, oldSpelling, newSpelling string, coherencyMap map[string]string) error {
	if oldSpelling == newSpelling {
		return fmt.Errorf("old and new spelling are the same in %s: %s", filePath, line)
	}
	if lookup, ok := coherencyMap[newSpelling]; ok && lookup == oldSpelling {
		return fmt.Errorf("contradictory entry in %s: '%s' suggests '%s' and vice versa", filePath, oldSpelling, lookup)
	}
	if prev, ok := coherencyMap[oldSpelling]; ok && prev != newSpelling {
		return fmt.Errorf("duplicate key in %s: %s, val: %s vs. %s", filePath, oldSpelling, prev, newSpelling)
	}
	return nil
}

// buildSpellingAhoTrie builds an Aho-Corasick automaton over UTF-16 units from dict keys.
func buildSpellingAhoTrie(dict map[string]string) *SpellingAhoTrie {
	root := &spellingACNode{next: map[uint16]*spellingACNode{}}
	for k, v := range dict {
		if k == "" {
			continue
		}
		ku := utf16.Encode([]rune(k))
		n := root
		for _, unit := range ku {
			ch, ok := n.next[unit]
			if !ok {
				ch = &spellingACNode{next: map[uint16]*spellingACNode{}}
				n.next[unit] = ch
			}
			n = ch
		}
		n.out = append(n.out, spellingACOut{lenU16: len(ku), value: v})
	}
	// BFS fail links
	q := make([]*spellingACNode, 0, 256)
	for _, ch := range root.next {
		ch.fail = root
		q = append(q, ch)
	}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		for b, ch := range n.next {
			q = append(q, ch)
			f := n.fail
			for f != nil && f.next[b] == nil {
				f = f.fail
			}
			if f == nil {
				ch.fail = root
			} else {
				ch.fail = f.next[b]
				if ch.fail == nil {
					ch.fail = root
				}
			}
			if ch.fail != nil && len(ch.fail.out) > 0 {
				ch.out = append(ch.out, ch.fail.out...)
			}
		}
	}
	return &SpellingAhoTrie{root: root, dict: dict}
}

// LoadSpellingDataBoth builds body + sentence-start SpellingData from CSV content string.
// expand is the synthesizeForPosTags hook (nil = no ß form expansion).
// Prefer NewSpellingData for production loading of official alt_neu.csv.
func LoadSpellingDataBoth(content, pathHint string, expand func(string) []string) (*SpellingData, error) {
	return newSpellingDataFromLines(splitResourceLines(content), pathHint, expand)
}

// LoadSpellingData parses "old;new" CSV lines without synthesizer expansion.
func LoadSpellingData(r io.Reader, pathHint string) (*SpellingData, error) {
	return LoadSpellingDataWithExpand(r, pathHint, nil)
}

// LoadSpellingDataWithExpand ports getCoherencyMap for body + sentence-start from a reader.
func LoadSpellingDataWithExpand(r io.Reader, pathHint string, expand func(string) []string) (*SpellingData, error) {
	if pathHint == "" {
		pathHint = "spelling data"
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return LoadSpellingDataBoth(string(data), pathHint, expand)
}
