package hunspell

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// LoadDicWords loads a Hunspell .dic-style word list (first line may be a count).
// Lines of the form "word/flags" keep only the word part. Comments (#) and empty lines skipped.
func LoadDicWords(r io.Reader) ([]string, error) {
	sc := bufio.NewScanner(r)
	var words []string
	first := true
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if first {
			first = false
			// optional word count header
			if _, err := strconv.Atoi(line); err == nil {
				continue
			}
		}
		if i := strings.IndexByte(line, '/'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line != "" {
			words = append(words, line)
		}
	}
	return words, sc.Err()
}

// NewMapHunspellDictionaryFromDic builds a map dictionary from a .dic reader.
func NewMapHunspellDictionaryFromDic(r io.Reader) (*MapHunspellDictionary, error) {
	words, err := LoadDicWords(r)
	if err != nil {
		return nil, err
	}
	return NewMapHunspellDictionary(words), nil
}
