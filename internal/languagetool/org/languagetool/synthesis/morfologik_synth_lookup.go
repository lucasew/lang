package synthesis

import (
	"bufio"
	"os"
	"strings"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MorfologikSynthLookup returns forms for lemma+posTag via a synth .dict
// (Java BaseSynthesizer.lookup: stemmer.lookup(lemma + "|" + posTag), stem = form).
func MorfologikSynthLookup(d *atticmorfo.Dictionary) func(lemma, posTag string) []string {
	if d == nil {
		return nil
	}
	return func(lemma, posTag string) []string {
		if lemma == "" || posTag == "" {
			return nil
		}
		// Java DictionaryLookup key for synthesizer dictionaries.
		forms, err := d.Lookup(lemma + "|" + posTag)
		if err != nil || len(forms) == 0 {
			return nil
		}
		out := make([]string, 0, len(forms))
		seen := map[string]struct{}{}
		for _, f := range forms {
			// Synth dict: surface form is stored as stem; tag often empty.
			s := f.Stem
			if s == "" {
				s = f.Tag
			}
			if s == "" {
				continue
			}
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
		return out
	}
}

// OpenMorfologikSynthLookup opens a .dict path for synthesizer Lookup.
func OpenMorfologikSynthLookup(dictPath string) (func(lemma, posTag string) []string, error) {
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil {
		return nil, err
	}
	return MorfologikSynthLookup(d), nil
}

// LoadPossibleTagsFile ports SynthesizerTools.loadWords for english_tags.txt.
func LoadPossibleTagsFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var tags []string
	seen := map[string]struct{}{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		// Java SynthesizerTools.loadWords: nextLine().trim() (String.trim).
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		tags = append(tags, line)
	}
	return tags, sc.Err()
}

// LoadManualSynthesizerFile opens a tab-separated form/lemma/pos file, or nil if missing.
func LoadManualSynthesizerFile(path string) (*ManualSynthesizer, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	return NewManualSynthesizer(f)
}
