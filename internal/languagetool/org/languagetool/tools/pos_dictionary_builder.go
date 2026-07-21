package tools

import (
	"fmt"
	"io"
	"strings"
)

// POSDictionaryBuilder ports org.languagetool.tools.POSDictionaryBuilder text pipeline.
// Binary Morfologik compile is deferred; this prepares/normalizes tab entries.
type POSDictionaryBuilder struct {
	*DictionaryBuilder
}

func NewPOSDictionaryBuilder(info map[string]string) *POSDictionaryBuilder {
	return &POSDictionaryBuilder{DictionaryBuilder: NewDictionaryBuilder(info)}
}

// NormalizeTaggerInput rewrites lines to wordform\tlemma\tpostag and drops empties.
func (b *POSDictionaryBuilder) NormalizeTaggerInput(r io.Reader, w io.Writer) (int, error) {
	entries, err := ReadTaggerEntries(r)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if e.Wordform == "" {
			continue
		}
		line := e.Wordform + "\t" + e.Lemma + "\t" + e.POSTag
		if _, err := fmt.Fprintln(w, line); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// Encoding returns fsa.dict.encoding from info props (default utf-8).
func (b *POSDictionaryBuilder) Encoding() string {
	if b == nil || b.Props == nil {
		return "utf-8"
	}
	if e := b.Props["fsa.dict.encoding"]; e != "" {
		return e
	}
	return "utf-8"
}

// Separator returns fsa.dict.separator if set (Java getOption → property.trim()).
func (b *POSDictionaryBuilder) Separator() string {
	if b == nil || b.Props == nil {
		return ""
	}
	return JavaStringTrim(b.Props["fsa.dict.separator"])
}

// ValidateTaggerLine checks tab format has at least a wordform.
func ValidateTaggerLine(line string) error {
	// Java DictionaryBuilder does not trim lines before split("\t").
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}
	if !strings.Contains(line, "\t") && strings.Contains(line, " ") {
		return fmt.Errorf("expected tab-separated fields, got spaces: %q", line)
	}
	return nil
}
