package synthesis

import (
	"bufio"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// LoadWords ports SynthesizerTools.loadWords — skip empty/# lines, intern forms.
func LoadWords(r io.Reader) ([]string, error) {
	var result []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		result = append(result, tools.Intern(line))
	}
	return result, sc.Err()
}
