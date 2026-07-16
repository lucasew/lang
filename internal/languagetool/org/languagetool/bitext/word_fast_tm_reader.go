package bitext

import (
	"fmt"
	"strings"
)

// WordFastTMReader ports org.languagetool.bitext.WordFastTMReader.
// WordFast TM files are tab-delimited; first line is a header skipped on open.
type WordFastTMReader struct {
	*TabBitextReader
}

func NewWordFastTMReader(filename, encoding string) (*WordFastTMReader, error) {
	r, err := newTabBitextReader(filename, wordFastParse)
	if err != nil {
		return nil, err
	}
	// skip header (first line already loaded as next) — re-read second line
	if r.nextLine != nil {
		if r.scanner.Scan() {
			line := r.scanner.Text()
			r.nextLine = &line
			p, err := wordFastParse(line)
			if err != nil {
				r.file.Close()
				return nil, err
			}
			r.nextPair = p
		} else {
			r.nextLine = nil
			r.nextPair = nil
		}
	}
	return &WordFastTMReader{TabBitextReader: r}, nil
}

func wordFastParse(line string) (*StringPair, error) {
	if line == "" {
		// empty line still invalid for WordFast data rows
	}
	fields := strings.Split(line, "\t")
	// Java: fields[4] source, fields[6] target
	if len(fields) < 7 {
		return nil, fmt.Errorf("Unexpected WordFast line format (need >= 7 fields): %s", line)
	}
	p := NewStringPair(fields[4], fields[6])
	return &p, nil
}
