package de

import (
	"os"
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
)

// GermanMultitokenSpeller ports org.languagetool.rules.de.GermanMultitokenSpeller.
// Java loads: /de/multitoken-suggest.txt, /spelling_global.txt, de/hunspell/spelling.txt
type GermanMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

// Resource paths used by the Java constructor (for loaders/discover).
var GermanMultitokenResourcePaths = []string{
	"/de/multitoken-suggest.txt",
	"/spelling_global.txt",
	"de/hunspell/spelling.txt",
}

// NewGermanMultitokenSpeller builds an empty speller with German isException + prepareLineForSpeller.
func NewGermanMultitokenSpeller() *GermanMultitokenSpeller {
	sp := multitoken.NewMultitokenSpeller()
	g := &GermanMultitokenSpeller{MultitokenSpeller: sp}
	sp.IsException = g.IsException
	// Java MultitokenSpeller uses language.prepareLineForSpeller on each dict line.
	sp.PrepareLine = PrepareLineForSpeller
	return g
}

// PrepareLineForSpeller ports German.prepareLineForSpeller.
// Line format: form[/E][/S][/N] with optional # comment. Expands form + e/s/n suffix tags.
// Example: "Haus/E" → ["Haus", "Hause"]; "Foo/ESN" → ["Foo", "Fooe", "Foos", "Foon"].
func PrepareLineForSpeller(line string) []string {
	// Java: String[] parts = line.split("#");
	parts := strings.Split(line, "#")
	if len(parts) == 0 {
		return []string{line}
	}
	// Java: parts[0].split("[/]")
	formTag := strings.Split(parts[0], "/")
	if len(formTag) == 0 {
		return []string{""}
	}
	form := formTag[0]
	results := []string{form}
	// Java: if (formTag.length==2) tag = formTag[1]; else tag stays "".
	// So "a/E/S" (length 3) does not expand — only exact two segments.
	tag := ""
	if len(formTag) == 2 {
		tag = formTag[1]
	}
	if strings.Contains(tag, "E") {
		results = append(results, form+"e")
	}
	if strings.Contains(tag, "S") {
		results = append(results, form+"s")
	}
	if strings.Contains(tag, "N") {
		results = append(results, form+"n")
	}
	return results
}

// INSTANCE mirrors the Java singleton for call sites that only need IsException.
var GermanMultitokenSpellerInstance = NewGermanMultitokenSpeller()

// IsException ports GermanMultitokenSpeller.isException:
// original without final UTF-16 unit 's' or '-' equals candidate
// (Java String.substring(0, length-1) / endsWith — UTF-16 indices).
func (g *GermanMultitokenSpeller) IsException(original, candidate string) bool {
	if original == "" || candidate == "" {
		return false
	}
	u := utf16.Encode([]rune(original))
	if len(u) < 1 {
		return false
	}
	// Java: original.substring(0, original.length()-1).equals(candidate)
	prefix := string(utf16.Decode(u[:len(u)-1]))
	if prefix != candidate {
		return false
	}
	// Java: original.endsWith("s") || original.endsWith("-")
	last := u[len(u)-1]
	return last == 's' || last == '-'
}

// LoadGermanMultitokenSpeller loads official resource files in Java constructor order.
// Empty paths are skipped. Returns loaded speller (possibly empty if no files open).
func LoadGermanMultitokenSpeller(paths ...string) (*GermanMultitokenSpeller, error) {
	s := NewGermanMultitokenSpeller()
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		err = s.LoadWords(f)
		_ = f.Close()
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// DiscoverAndLoadGermanMultitokenSpeller loads from DiscoverGermanResourceDir when present.
func DiscoverAndLoadGermanMultitokenSpeller() *GermanMultitokenSpeller {
	root := DiscoverGermanResourceDir()
	if root == "" {
		return NewGermanMultitokenSpeller()
	}
	// Map Java classpath paths to resource tree files.
	var paths []string
	for _, rel := range []string{
		"multitoken-suggest.txt",
		"spelling_global.txt", // may be under shared spelling/
		"hunspell/spelling.txt",
	} {
		// try under root and common parents
		for _, p := range []string{
			root + "/" + rel,
			root + "/../" + rel,
			root + "/../../" + rel,
		} {
			if fileExists(p) {
				paths = append(paths, p)
				break
			}
		}
	}
	// also spelling_global next to resource root
	if p := findSpellingGlobalNear(root); p != "" {
		paths = appendUniquePath(paths, p)
	}
	s, err := LoadGermanMultitokenSpeller(paths...)
	if err != nil || s == nil {
		return NewGermanMultitokenSpeller()
	}
	return s
}

func appendUniquePath(paths []string, p string) []string {
	for _, x := range paths {
		if x == p {
			return paths
		}
	}
	return append(paths, p)
}

func findSpellingGlobalNear(resourceRoot string) string {
	// Java /spelling_global.txt is often under org/languagetool/resource/
	cands := []string{
		resourceRoot + "/spelling_global.txt",
		resourceRoot + "/../spelling_global.txt",
		resourceRoot + "/../../spelling_global.txt",
	}
	for _, p := range cands {
		if fileExists(p) {
			return p
		}
	}
	return ""
}
