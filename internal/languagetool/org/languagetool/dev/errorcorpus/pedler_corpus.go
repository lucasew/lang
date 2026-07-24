package errorcorpus

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	normalizeERR = regexp.MustCompile(`(?i)\s*<ERR\s+targ\s*=\s*([^>]*?)\s*>\s*(.*?)\s*</ERR>\s*`)
	spacesRE     = regexp.MustCompile(`\s+`)
	stripERRRE   = regexp.MustCompile(`(?i)<ERR\s+targ=[^>]*>(.*?)</ERR>`)
)

// PedlerCorpus ports org.languagetool.dev.errorcorpus.PedlerCorpus — loads *.txt lines from a directory.
type PedlerCorpus struct {
	lines []string
	pos   int
}

// NewPedlerCorpus loads all *.txt files in dir (non-recursive).
func NewPedlerCorpus(dir string) (*PedlerCorpus, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		sc := bufio.NewScanner(f)
		// long lines
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			lines = append(lines, sc.Text())
		}
		_ = f.Close()
		if err := sc.Err(); err != nil {
			return nil, err
		}
	}
	return &PedlerCorpus{lines: lines}, nil
}

// HasNext reports remaining sentences.
func (c *PedlerCorpus) HasNext() bool {
	return c != nil && c.pos < len(c.lines)
}

// Next returns the next ErrorSentence.
func (c *PedlerCorpus) Next() *ErrorSentence {
	if !c.HasNext() {
		return nil
	}
	line := c.lines[c.pos]
	c.pos++
	return parsePedlerLine(line)
}

func parsePedlerLine(line string) *ErrorSentence {
	normalized := normalizeERR.ReplaceAllString(line, " <ERR targ=$1>$2</ERR> ")
	normalized = spacesRE.ReplaceAllString(normalized, " ")
	normalized = strings.TrimSpace(normalized)

	var errors []Error
	startPos := 0
	for {
		rel := strings.Index(normalized[startPos:], "<ERR targ=")
		if rel < 0 {
			break
		}
		startTagStart := startPos + rel
		startTagEndRel := strings.Index(normalized[startTagStart:], ">")
		if startTagEndRel < 0 {
			break
		}
		startTagEnd := startTagStart + startTagEndRel
		endTagRel := strings.Index(normalized[startTagStart:], "</ERR>")
		if endTagRel < 0 {
			break
		}
		endTagStart := startTagStart + endTagRel
		correction := normalized[startTagStart+len("<ERR targ=") : startTagEnd]
		errors = append(errors, Error{
			StartPos:   startTagEnd + 1,
			EndPos:     endTagStart,
			Correction: correction,
		})
		startPos = startTagStart + 1
	}
	plain := stripERRRE.ReplaceAllString(normalized, "$1")
	plain = spacesRE.ReplaceAllString(plain, " ")
	plain = strings.TrimSpace(plain)
	return &ErrorSentence{MarkupText: normalized, PlainText: plain, Errors: errors}
}
