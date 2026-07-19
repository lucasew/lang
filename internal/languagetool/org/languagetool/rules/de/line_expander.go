package de

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// LineExpander ports org.languagetool.rules.de.LineExpander.
// VerbForms optional: Java GermanSynthesizer.synthesizeForPosTags(verb, VER:*).
// Without VerbForms, verb-prefix lines still emit join, zu-form, and genitive (always
// present in Java after synth forms); full conjugation needs synthesizer resources.
type LineExpander struct {
	// VerbForms returns conjugated forms for a verb lemma (VER: tags only).
	VerbForms func(lemma string) []string
}

func NewLineExpander() *LineExpander { return &LineExpander{} }

var (
	lineExpanderOnce sync.Once
	wiredLineExp     *LineExpander
)

// WireLineExpander returns a LineExpander with VerbForms from german_synth.dict when present.
// Java: GermanSynthesizer.INSTANCE.synthesizeForPosTags(verb, s -> s.startsWith("VER:")).
func WireLineExpander() *LineExpander {
	lineExpanderOnce.Do(func() {
		wiredLineExp = NewLineExpander()
		if gs := openDiscoveredGermanSynthesizer(); gs != nil {
			wiredLineExp.VerbForms = func(lemma string) []string {
				if lemma == "" {
					return nil
				}
				return gs.SynthesizeForPosTags(lemma, func(tag string) bool {
					return strings.HasPrefix(tag, "VER:")
				})
			}
			return
		}
		if base := openDiscoveredGermanSynthBase(); base != nil {
			wiredLineExp.VerbForms = func(lemma string) []string {
				if lemma == "" {
					return nil
				}
				return base.SynthesizeForPosTags(lemma, func(tag string) bool {
					return strings.HasPrefix(tag, "VER:")
				})
			}
		}
	})
	if wiredLineExp == nil {
		return NewLineExpander()
	}
	cp := *wiredLineExp
	return &cp
}

func (e *LineExpander) ExpandLine(line string) []string {
	if isLineWithVerbPrefix(line) {
		return e.handleLineWithPrefix(line)
	}
	if isLineWithFlag(line) {
		return handleLineWithFlags(line)
	}
	// Java: including "" and "#comment" → cleaned singleton list
	return []string{cleanTagsAndEscapeChars(line)}
}

func isLineWithFlag(line string) bool {
	idx := strings.IndexByte(line, '/')
	return !strings.HasPrefix(line, "#") && idx > 0 && line[idx-1] != '\\'
}

func isLineWithVerbPrefix(line string) bool {
	idx := strings.IndexByte(line, '_')
	return !strings.HasPrefix(line, "#") && idx > 0 && line[idx-1] != '\\'
}

func (e *LineExpander) handleLineWithPrefix(line string) []string {
	clean := cleanTagsAndEscapeChars(line)
	parts := strings.Split(clean, "_")
	if len(parts) != 2 {
		panic(fmt.Sprintf("Unexpected line format, expected at most one '_': %s", line))
	}
	if strings.Contains(parts[0], "/") || strings.Contains(parts[1], "/") {
		panic(fmt.Sprintf("Unexpected line format, '_' cannot be combined with '/': %s", line))
	}
	// Gender-gap: Lehrer_in
	if parts[1] == "in" {
		p0 := parts[0]
		return []string{
			p0 + "_in",
			p0 + "_innen",
			p0 + "*in",
			p0 + "*innen",
			p0 + ":in",
			p0 + ":innen",
		}
	}
	var result []string
	seen := map[string]struct{}{}
	add := func(w string) {
		if _, ok := seen[w]; ok {
			return
		}
		seen[w] = struct{}{}
		result = append(result, w)
	}
	// Verb forms from synthesizer when available
	var forms []string
	if e != nil && e.VerbForms != nil {
		forms = e.VerbForms(parts[1])
	}
	if len(forms) == 0 && e != nil && e.VerbForms != nil {
		// Java throws if synthesizer returns empty
		panic(fmt.Sprintf("Could not expand '%s' from line '%s', no forms found", parts[1], line))
	}
	for _, form := range forms {
		if form == "" {
			continue
		}
		// skip ß forms and non-lowercase starts (old spellings risk)
		if strings.Contains(form, "ß") {
			continue
		}
		r, _ := utf8DecodeFirst(form)
		if !unicode.IsLower(r) {
			continue
		}
		add(parts[0] + form)
	}
	// Always: zu-verb and genitive (Java adds even after synth forms)
	add(parts[0] + "zu" + parts[1])
	add(tools.UppercaseFirstChar(parts[0]) + parts[1] + "s")
	// Without synthesizer, also emit plain join so callers still get a usable form
	if len(forms) == 0 {
		add(parts[0] + parts[1])
	}
	return result
}

func utf8DecodeFirst(s string) (rune, int) {
	for _, r := range s {
		return r, 1
	}
	return 0, 0
}

func handleLineWithFlags(line string) []string {
	clean := cleanTagsAndEscapeChars(line)
	parts := strings.SplitN(clean, "/", 2)
	if len(parts) != 2 {
		panic(fmt.Sprintf("Unexpected line format, expected at most one slash: %s", line))
	}
	word, suffix := parts[0], parts[1]
	var result []string
	add := func(w string) {
		for _, x := range result {
			if x == w {
				return
			}
		}
		result = append(result, w)
	}
	for _, c := range suffix {
		switch c {
		case 'S':
			add(word)
			add(word + "s")
		case 'N':
			add(word)
			add(word + "n")
		case 'E':
			add(word)
			add(word + "e")
		case 'F':
			add(word)
			add(word + "in")
		case 'T':
			add(word)
			if strings.HasSuffix(word, "straße") || strings.HasSuffix(word, "strasse") {
				// Java replaceAll("stra(ß|ss)e", "str.")
				w := word
				w = strings.ReplaceAll(w, "straße", "str.")
				w = strings.ReplaceAll(w, "strasse", "str.")
				add(w)
			}
			if strings.HasSuffix(word, "Straße") || strings.HasSuffix(word, "Strasse") {
				w := word
				w = strings.ReplaceAll(w, "Straße", "Str.")
				w = strings.ReplaceAll(w, "Strasse", "Str.")
				add(w)
			}
		case 'A', 'P':
			add(word)
			if strings.HasSuffix(word, "e") {
				add(word + "r")
				add(word + "s")
				add(word + "n")
				add(word + "m")
			} else {
				add(word + "e")
				add(word + "er")
				add(word + "es")
				add(word + "en")
				add(word + "em")
			}
		default:
			panic(fmt.Sprintf("Unknown suffix: %s in line: %s", suffix, line))
		}
	}
	if len(result) == 0 {
		return []string{word}
	}
	return result
}

func cleanTagsAndEscapeChars(s string) string {
	if idx := strings.IndexByte(s, '#'); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(strings.ReplaceAll(s, "\\", ""))
}
