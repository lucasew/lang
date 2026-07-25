package de

import (
	"fmt"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// LineExpander ports org.languagetool.rules.de.LineExpander.
// VerbForms: Java GermanSynthesizer.INSTANCE.synthesizeForPosTags(verb, s -> s.startsWith("VER:")).
// When german_synth.dict is missing VerbForms is nil → verb-prefix lines fail-closed
// (no invent join/zu/genitive). Empty non-nil VerbForms panics like Java RuntimeException.
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

// ExpandLine ports LineExpander.expandLine.
func (e *LineExpander) ExpandLine(line string) []string {
	if isLineWithVerbPrefix(line) {
		return e.handleLineWithPrefix(line)
	}
	if isLineWithFlag(line) {
		return handleLineWithFlags(line)
	}
	// Java: Collections.singletonList(cleanTagsAndEscapeChars(line))
	return []string{cleanTagsAndEscapeChars(line)}
}

func isLineWithFlag(line string) bool {
	// Java: indexOf('/') then charAt(idx-1) — '/' and '\\' are BMP/ASCII
	idx := strings.IndexByte(line, '/')
	return !strings.HasPrefix(line, "#") && idx > 0 && line[idx-1] != '\\'
}

func isLineWithVerbPrefix(line string) bool {
	// Java: indexOf('_') then charAt(idx-1)
	idx := strings.IndexByte(line, '_')
	return !strings.HasPrefix(line, "#") && idx > 0 && line[idx-1] != '\\'
}

// javaSplitLimit0 ports Java String.split(sep) with limit 0 (trailing empties discarded).
func javaSplitLimit0(s, sep string) []string {
	parts := strings.Split(s, sep)
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func (e *LineExpander) handleLineWithPrefix(line string) []string {
	// Java: cleanTagsAndEscapeChars(line).split("_")
	parts := javaSplitLimit0(cleanTagsAndEscapeChars(line), "_")
	if len(parts) != 2 {
		panic(fmt.Sprintf("Unexpected line format, expected at most one '_': %s", line))
	}
	if strings.Contains(parts[0], "/") || strings.Contains(parts[1], "/") {
		panic(fmt.Sprintf("Unexpected line format, '_' cannot be combined with '/': %s", line))
	}
	// Gender-gap special case: …_in
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
	// Java: verbFormCache.getUnchecked(parts[1]); if forms.length == 0 → RuntimeException
	// Never invent plain join/zu/genitive without synthesizer forms.
	if e == nil || e.VerbForms == nil {
		// Synthesizer asset missing: fail-closed (no invent). Java always has INSTANCE.
		return nil
	}
	forms := e.VerbForms(parts[1])
	if len(forms) == 0 {
		// Java: throw new RuntimeException("Could not expand ... no forms found")
		panic(fmt.Sprintf("Could not expand '%s' from line '%s', no forms found", parts[1], line))
	}
	// Java: Set<String> formSet = new HashSet<>(Arrays.asList(forms));
	formSet := make(map[string]struct{}, len(forms))
	for _, form := range forms {
		formSet[form] = struct{}{}
	}
	var result []string
	for form := range formSet {
		// Java: !form.contains("ß") && form.length() > 0 && Character.isLowerCase(form.charAt(0))
		if form == "" || strings.Contains(form, "ß") {
			continue
		}
		// Character.isLowerCase(form.charAt(0)) ≈ tools.StartsWithLowercase (UTF-16 unit 0)
		if !tools.StartsWithLowercase(form) {
			continue
		}
		// skip old spellings with ß; only lower-case synth forms
		result = append(result, parts[0]+form)
	}
	// Java always appends after non-empty forms:
	// parts[0] + "zu" + parts[1]
	// uppercaseFirstChar(parts[0]) + parts[1] + "s"
	result = append(result, parts[0]+"zu"+parts[1])
	result = append(result, tools.UppercaseFirstChar(parts[0])+parts[1]+"s")
	return result
}

func handleLineWithFlags(line string) []string {
	// Java: cleanTagsAndEscapeChars(line).split("/") — at most one slash
	parts := javaSplitLimit0(cleanTagsAndEscapeChars(line), "/")
	if len(parts) != 2 {
		panic(fmt.Sprintf("Unexpected line format, expected at most one slash: %s", line))
	}
	word, suffix := parts[0], parts[1]
	var result []string
	add := func(w string) {
		// Java private add(): if (!result.contains(word)) result.add(word)
		for _, x := range result {
			if x == w {
				return
			}
		}
		result = append(result, w)
	}
	// Java: for (int i = 0; i < suffix.length(); i++) { char c = suffix.charAt(i); ... }
	for i := 0; i < utf16LenDE(suffix); i++ {
		c := javaCharAtDE(suffix, i)
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
			add(word + "in") // (m/f)
		case 'T':
			add(word)
			if strings.HasSuffix(word, "straße") || strings.HasSuffix(word, "strasse") {
				// Java: word.replaceAll("stra(ß|ss)e", "str.")
				w := word
				w = strings.ReplaceAll(w, "straße", "str.")
				w = strings.ReplaceAll(w, "strasse", "str.")
				add(w)
			}
			if strings.HasSuffix(word, "Straße") || strings.HasSuffix(word, "Strasse") {
				// Java: word.replaceAll("Stra(ß|ss)e", "Str.")
				w := word
				w = strings.ReplaceAll(w, "Straße", "Str.")
				w = strings.ReplaceAll(w, "Strasse", "Str.")
				add(w)
			}
		case 'A', 'P': // Adjektiv / Partizip
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
			// Java: throw new IllegalArgumentException("Unknown suffix: " + suffix + " in line: " + line)
			panic(fmt.Sprintf("Unknown suffix: %s in line: %s", suffix, line))
		}
	}
	return result
}

// cleanTagsAndEscapeChars ports LineExpander.cleanTagsAndEscapeChars:
// ignore "#..." tags; strip escape backslashes; trim.
func cleanTagsAndEscapeChars(s string) string {
	if idx := strings.IndexByte(s, '#'); idx >= 0 {
		s = s[:idx]
	}
	return tools.JavaStringTrim(strings.ReplaceAll(s, "\\", ""))
}
