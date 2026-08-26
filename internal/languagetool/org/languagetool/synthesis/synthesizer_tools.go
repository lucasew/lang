package synthesis

import (
	"bufio"
	"io"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// LoadWords ports SynthesizerTools.loadWords — skip empty/# lines, intern forms.
// Java: line = scanner.nextLine().trim() (String.trim, not Unicode TrimSpace).
func LoadWords(r io.Reader) ([]string, error) {
	var result []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		result = append(result, tools.Intern(line))
	}
	return result, sc.Err()
}

// SynthesizerTools ports org.languagetool.synthesis.SynthesizerTools.
type SynthesizerTools struct{}

// LoadWords loads non-comment lines (same as package LoadWords).
func (SynthesizerTools) LoadWords(r io.Reader) ([]string, error) {
	return LoadWords(r)
}
