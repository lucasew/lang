package dev

import (
	"bufio"
	"io"
)

const DefaultSkipThreshold = 0.95

// DetectedLang is a pluggable language-id result.
type DetectedLang struct {
	ShortCode  string
	Confidence float64
}

// DetectLangFunc detects language for a line.
type DetectLangFunc func(line string) *DetectedLang

// FilterFileByLanguage ports org.languagetool.dev.FilterFileByLanguage:
// writes lines that are not confidently a different language than expected.
// Returns skip count.
func FilterFileByLanguage(r io.Reader, w io.Writer, expectedLang string, detect DetectLangFunc, skipThreshold float64) (int, error) {
	if skipThreshold <= 0 {
		skipThreshold = DefaultSkipThreshold
	}
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	skip := 0
	for sc.Scan() {
		line := sc.Text()
		if detect != nil {
			if d := detect(line); d != nil &&
				d.ShortCode != expectedLang &&
				d.Confidence > skipThreshold {
				skip++
				continue
			}
		}
		if _, err := io.WriteString(w, line+"\n"); err != nil {
			return skip, err
		}
	}
	return skip, sc.Err()
}
