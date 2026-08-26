package tools

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	freqRangesIn     = 256
	freqRangesOut    = 26 // A-Z
	firstRangeCode   = 65 // 'A'
)

// DictionaryBuilder ports the text-side helpers of org.languagetool.tools.DictionaryBuilder
// (binary FSA compile is deferred).
type DictionaryBuilder struct {
	Props      map[string]string
	OutputFile string
	FreqList   map[string]int
}

func NewDictionaryBuilder(infoProps map[string]string) *DictionaryBuilder {
	if infoProps == nil {
		infoProps = map[string]string{}
	}
	return &DictionaryBuilder{
		Props:    infoProps,
		FreqList: map[string]int{},
	}
}

func (b *DictionaryBuilder) SetOutputFilename(name string) {
	if b != nil {
		b.OutputFile = name
	}
}

func (b *DictionaryBuilder) GetOutputFilename() string {
	if b == nil {
		return ""
	}
	return b.OutputFile
}

// LoadFrequencyList parses a simple frequency XML subset: <w f="N">word</w>
func (b *DictionaryBuilder) LoadFrequencyList(r io.Reader) error {
	if b == nil {
		return nil
	}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		// crude extract f=".." and >word<
		fIdx := strings.Index(line, `f="`)
		if fIdx < 0 {
			continue
		}
		rest := line[fIdx+3:]
		endQ := strings.IndexByte(rest, '"')
		if endQ < 0 {
			continue
		}
		freq, err := strconv.Atoi(rest[:endQ])
		if err != nil {
			continue
		}
		gt := strings.IndexByte(line, '>')
		lt := strings.LastIndexByte(line, '<')
		if gt < 0 || lt <= gt {
			continue
		}
		word := line[gt+1 : lt]
		if word != "" {
			b.FreqList[word] = freq
		}
	}
	return sc.Err()
}

// FreqToRange maps an input frequency (0..255 style) to A-Z range code.
func FreqToRange(freq int) byte {
	if freq < 0 {
		freq = 0
	}
	if freq >= freqRangesIn {
		freq = freqRangesIn - 1
	}
	// map to 0..25
	bucket := freq * freqRangesOut / freqRangesIn
	if bucket >= freqRangesOut {
		bucket = freqRangesOut - 1
	}
	return byte(firstRangeCode + bucket)
}

// TaggerEntry is one wordform/lemma/postag line.
type TaggerEntry struct {
	Wordform string
	Lemma    string
	POSTag   string
}

// ParseTaggerLine parses wordform\tlemma\tpostag (or spelling-only wordform).
// Java DictionaryBuilder: line.split("\t") with no trim; ignore unless length==3 for POS export.
// Here we accept ≥1 tab fields for tooling; do not invent Unicode TrimSpace.
func ParseTaggerLine(line string) (TaggerEntry, bool) {
	if line == "" {
		return TaggerEntry{}, false
	}
	// Soft skip comments for human-edited files (not in Java path for raw dict build).
	if strings.HasPrefix(line, "#") {
		return TaggerEntry{}, false
	}
	parts := strings.Split(line, "\t")
	e := TaggerEntry{Wordform: parts[0]}
	if len(parts) > 1 {
		e.Lemma = parts[1]
	}
	if len(parts) > 2 {
		e.POSTag = parts[2]
	}
	return e, true
}

// ReadTaggerEntries reads tab-separated dictionary lines.
func ReadTaggerEntries(r io.Reader) ([]TaggerEntry, error) {
	var out []TaggerEntry
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		if e, ok := ParseTaggerLine(sc.Text()); ok {
			out = append(out, e)
		}
	}
	return out, sc.Err()
}

// WriteSpellingList writes unique wordforms (one per line) for spell dictionaries.
func WriteSpellingList(w io.Writer, entries []TaggerEntry) error {
	seen := map[string]struct{}{}
	for _, e := range entries {
		if e.Wordform == "" {
			continue
		}
		if _, ok := seen[e.Wordform]; ok {
			continue
		}
		seen[e.Wordform] = struct{}{}
		if _, err := fmt.Fprintln(w, e.Wordform); err != nil {
			return err
		}
	}
	return nil
}

// Separator returns fsa.dict.separator from props.
// Java DictionaryBuilder.getOption: return property.trim().
func (b *DictionaryBuilder) Separator() string {
	if b == nil || b.Props == nil {
		return ""
	}
	return JavaStringTrim(b.Props["fsa.dict.separator"])
}
