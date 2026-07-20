package bitext

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf16"
)

// pairParser converts a line into a StringPair (nil line → nil pair).
type pairParser func(line string) (*StringPair, error)

// TabBitextReader ports org.languagetool.bitext.TabBitextReader.
type TabBitextReader struct {
	file        *os.File
	scanner     *bufio.Scanner
	nextLine    *string
	nextPair    *StringPair
	prevLine    string
	lineCount   int
	sentencePos int
	closed      bool
	parse       pairParser
}

func NewTabBitextReader(filename, encoding string) (*TabBitextReader, error) {
	return newTabBitextReader(filename, defaultTabParse)
}

func newTabBitextReader(filename string, parse pairParser) (*TabBitextReader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	r := &TabBitextReader{
		file:      f,
		scanner:   bufio.NewScanner(f),
		lineCount: -1,
		parse:     parse,
	}
	if r.scanner.Scan() {
		line := r.scanner.Text()
		r.nextLine = &line
		p, err := r.parse(line)
		if err != nil {
			f.Close()
			return nil, err
		}
		r.nextPair = p
	}
	r.prevLine = ""
	return r, nil
}

func defaultTabParse(line string) (*StringPair, error) {
	// Java tab2StringPair: null line → null; else split on tab
	fields := strings.Split(line, "\t")
	if len(fields) < 2 {
		return nil, fmt.Errorf("Unexpected format, expected two tab-separated columns: %s", line)
	}
	p := NewStringPair(fields[0], fields[1])
	return &p, nil
}

func (r *TabBitextReader) HasNext() bool {
	return r.nextLine != nil
}

func (r *TabBitextReader) Next() (StringPair, bool, error) {
	// Ports TabBitextReader.TabReader.next
	if r.nextLine == nil || r.nextPair == nil {
		return StringPair{}, false, nil
	}
	result := *r.nextPair
	// Java String.length() is UTF-16 code units
	r.sentencePos = utf16LenBitext(result.GetSource()) + 1
	r.prevLine = *r.nextLine
	if r.scanner.Scan() {
		line := r.scanner.Text()
		r.nextLine = &line
		p, err := r.parse(line)
		if err != nil {
			return result, true, err
		}
		r.nextPair = p
		r.lineCount++
	} else {
		r.nextLine = nil
		r.nextPair = nil
		r.lineCount++
		if !r.closed {
			_ = r.file.Close()
			r.closed = true
		}
	}
	return result, true, nil
}

func (r *TabBitextReader) GetColumnCount() int       { return r.sentencePos }
func (r *TabBitextReader) GetTargetColumnCount() int { return 1 }
func (r *TabBitextReader) GetLineCount() int         { return r.lineCount }
func (r *TabBitextReader) GetSentencePosition() int  { return r.sentencePos }
func (r *TabBitextReader) GetCurrentLine() string    { return r.prevLine }

func utf16LenBitext(s string) int {
	return len(utf16.Encode([]rune(s)))
}
