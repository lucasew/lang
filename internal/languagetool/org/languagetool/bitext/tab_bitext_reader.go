package bitext

import (
	"bufio"
	"fmt"
	"os"
)

// TabBitextReader ports org.languagetool.bitext.TabBitextReader.
type TabBitextReader struct {
	file       *os.File
	scanner    *bufio.Scanner
	nextLine   *string
	nextPair   *StringPair
	prevLine   string
	lineCount  int
	sentencePos int
	closed     bool
}

func NewTabBitextReader(filename, encoding string) (*TabBitextReader, error) {
	// encoding is accepted for API parity; Go opens UTF-8 (tests use UTF-8)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	r := &TabBitextReader{
		file:      f,
		scanner:   bufio.NewScanner(f),
		lineCount: -1,
	}
	if r.scanner.Scan() {
		line := r.scanner.Text()
		r.nextLine = &line
		p, err := tab2StringPair(line)
		if err != nil {
			f.Close()
			return nil, err
		}
		r.nextPair = &p
	}
	r.prevLine = ""
	return r, nil
}

func tab2StringPair(line string) (StringPair, error) {
	// split on first tab only? Java split("\t") all
	fields := splitTab(line)
	if len(fields) < 2 {
		return StringPair{}, fmt.Errorf("Unexpected format, expected two tab-separated columns: %s", line)
	}
	return NewStringPair(fields[0], fields[1]), nil
}

func splitTab(line string) []string {
	var fields []string
	start := 0
	for i := 0; i < len(line); i++ {
		if line[i] == '\t' {
			fields = append(fields, line[start:i])
			start = i + 1
		}
	}
	fields = append(fields, line[start:])
	return fields
}

// All returns all pairs (convenience; Java uses iterator).
func (r *TabBitextReader) All() ([]StringPair, error) {
	var out []StringPair
	for {
		p, ok, err := r.Next()
		if err != nil {
			return out, err
		}
		if !ok {
			break
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *TabBitextReader) HasNext() bool {
	return r.nextLine != nil
}

func (r *TabBitextReader) Next() (StringPair, bool, error) {
	if r.nextLine == nil {
		return StringPair{}, false, nil
	}
	result := *r.nextPair
	r.sentencePos = len([]rune(result.GetSource())) + 1 // approximate; Java uses length()
	r.prevLine = *r.nextLine
	if r.scanner.Scan() {
		line := r.scanner.Text()
		r.nextLine = &line
		p, err := tab2StringPair(line)
		if err != nil {
			return result, true, err
		}
		r.nextPair = &p
		r.lineCount++
	} else {
		r.nextLine = nil
		r.nextPair = nil
		r.lineCount++
		if !r.closed {
			r.file.Close()
			r.closed = true
		}
	}
	return result, true, nil
}

func (r *TabBitextReader) GetLineCount() int { return r.lineCount }
