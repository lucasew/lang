package tools

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// PolishIgnoreRegex ports SynthDictionaryBuilder.POLISH_IGNORE_REGEX default.
const PolishIgnorePOSRegex = ":neg|qub|depr"

// SynthDictionaryBuilder ports org.languagetool.tools.SynthDictionaryBuilder text side.
// Synthesis dict stores lemma\twordform\tpostag (reversed from tagger order).
type SynthDictionaryBuilder struct {
	*DictionaryBuilder
	IgnorePOS *regexp.Regexp
	IgnoreItems map[string]struct{}
}

func NewSynthDictionaryBuilder(info map[string]string) *SynthDictionaryBuilder {
	return &SynthDictionaryBuilder{
		DictionaryBuilder: NewDictionaryBuilder(info),
		IgnoreItems:       map[string]struct{}{},
	}
}

// SetIgnorePOSRegex compiles a POS filter (forms matching are dropped).
func (b *SynthDictionaryBuilder) SetIgnorePOSRegex(pattern string) error {
	if pattern == "" {
		b.IgnorePOS = nil
		return nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	b.IgnorePOS = re
	return nil
}

// ReverseLineContent converts tagger lines to synth order: lemma\twordform\tpostag.
func (b *SynthDictionaryBuilder) ReverseLineContent(r io.Reader, w io.Writer) (int, error) {
	entries, err := ReadTaggerEntries(r)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if e.Wordform == "" || e.Lemma == "" {
			continue
		}
		if _, bad := b.IgnoreItems[e.Wordform]; bad {
			continue
		}
		if b.IgnorePOS != nil && b.IgnorePOS.MatchString(e.POSTag) {
			continue
		}
		line := e.Lemma + "\t" + e.Wordform + "\t" + e.POSTag
		if _, err := fmt.Fprintln(w, line); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// WritePOSTags writes unique POS tags (one per line) for *_tags.txt.
func WritePOSTags(r io.Reader, w io.Writer) (int, error) {
	entries, err := ReadTaggerEntries(r)
	if err != nil {
		return 0, err
	}
	seen := map[string]struct{}{}
	n := 0
	for _, e := range entries {
		if e.POSTag == "" {
			continue
		}
		if _, ok := seen[e.POSTag]; ok {
			continue
		}
		seen[e.POSTag] = struct{}{}
		if _, err := fmt.Fprintln(w, e.POSTag); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// LoadIgnoreItems loads one-word-per-line ignore list (filter-archaic.txt style).
func (b *SynthDictionaryBuilder) LoadIgnoreItems(r io.Reader) error {
	if b.IgnoreItems == nil {
		b.IgnoreItems = map[string]struct{}{}
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		b.IgnoreItems[line] = struct{}{}
	}
	return nil
}
