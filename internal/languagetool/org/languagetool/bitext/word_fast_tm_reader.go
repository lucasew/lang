package bitext

import (
	"fmt"
	"strings"
	"unicode/utf16"
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
	// (Java WordFastTMReader ctor after super: nextLine = in.readLine(); nextPair = tab2StringPair(nextLine))
	if r.nextLine != nil {
		if r.scanner.Scan() {
			line := r.scanner.Text()
			r.nextLine = &line
			p, pos, err := wordFastParseWithPos(line)
			if err != nil {
				r.file.Close()
				return nil, err
			}
			r.nextPair = p
			r.sentencePos = pos
		} else {
			r.nextLine = nil
			r.nextPair = nil
		}
	}
	return &WordFastTMReader{TabBitextReader: r}, nil
}

func wordFastParse(line string) (*StringPair, error) {
	p, _, err := wordFastParseWithPos(line)
	return p, err
}

// wordFastParseWithPos ports WordFastTMReader.tab2StringPair:
// fields[4] source, fields[6] target; sentencePos = fields[4].length() + 1.
func wordFastParseWithPos(line string) (*StringPair, int, error) {
	fields := strings.Split(line, "\t")
	if len(fields) < 7 {
		return nil, 0, fmt.Errorf("Unexpected WordFast line format (need >= 7 fields): %s", line)
	}
	p := NewStringPair(fields[4], fields[6])
	// Java String.length() is UTF-16 code units
	pos := len(utf16.Encode([]rune(fields[4]))) + 1
	return &p, pos, nil
}

// Next ports WordFastTMReader.TabReader.next (does not set sentencePos from the
// returned pair — tab2StringPair on the following line updates sentencePos).
func (w *WordFastTMReader) Next() (StringPair, bool, error) {
	r := w.TabBitextReader
	if r.nextLine == nil || r.nextPair == nil {
		return StringPair{}, false, nil
	}
	result := *r.nextPair
	if r.scanner.Scan() {
		line := r.scanner.Text()
		r.nextLine = &line
		p, pos, err := wordFastParseWithPos(line)
		if err != nil {
			return result, true, err
		}
		r.nextPair = p
		r.sentencePos = pos
	} else {
		r.nextLine = nil
		r.nextPair = nil
		if !r.closed {
			_ = r.file.Close()
			r.closed = true
		}
	}
	return result, true, nil
}
