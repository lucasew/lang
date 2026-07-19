package de

import (
	"bufio"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PrefixInfixVerb ports GermanTagger.PrefixInfixVerb (spelling.txt "prefix_verb").
type PrefixInfixVerb struct {
	Prefix       string
	Infix        string // "" or "zu"
	VerbBaseform string
}

// SpellingVerbExpansion ports GermanTagger ExpansionInfos verb / nominalized maps
// from spelling.txt lines "prefix_verbbase" (not ending _in).
// Full conjugated surfaces need a synthesizer; without it we still register:
//   prefix+zu+base (VER:EIZ), Nominalized (SUB:…:INF), Genitive +s.
// verbInfos also maps prefix+base (infinitive surface) for prefix stripping in GermanTagger.
type SpellingVerbExpansion struct {
	// VerbInfos: full surface → prefix/infix/base (Java verbInfos)
	VerbInfos map[string]PrefixInfixVerb
	// fixed readings for nominalized / zu forms (WordTagger surface)
	byForm map[string][]tagging.TaggedWord
}

// LoadSpellingVerbExpansionFromFile loads underscore verb expansions.
func LoadSpellingVerbExpansionFromFile(path string) (*SpellingVerbExpansion, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadSpellingVerbExpansion(f)
}

// LoadSpellingVerbExpansion ports the underscore branch of initExpansionInfos.
// Optional synth: when non-nil, called as synth(verbBase) → verb forms (VER:…);
// forms containing "ß" are skipped (Java).
func LoadSpellingVerbExpansion(r io.Reader) (*SpellingVerbExpansion, error) {
	return LoadSpellingVerbExpansionWithSynth(r, nil)
}

// VerbFormSynth synthesizes VER: surfaces for a verb base (optional).
type VerbFormSynth func(verbBase string) []string

// LoadSpellingVerbExpansionWithSynth is LoadSpellingVerbExpansion plus optional conjugation list.
func LoadSpellingVerbExpansionWithSynth(r io.Reader, synth VerbFormSynth) (*SpellingVerbExpansion, error) {
	ex := &SpellingVerbExpansion{
		VerbInfos: map[string]PrefixInfixVerb{},
		byForm:    map[string][]tagging.TaggedWord{},
	}
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" || !strings.Contains(line, "_") || strings.HasSuffix(line, "_in") {
			continue
		}
		// flags like /S not combined with underscore in Java LineExpander; strip flags if any
		if i := strings.IndexByte(line, '/'); i >= 0 {
			line = line[:i]
		}
		parts := strings.Split(line, "_")
		if len(parts) != 2 {
			continue
		}
		prefix, verbBase := parts[0], parts[1]
		if prefix == "" || verbBase == "" {
			continue
		}
		if strings.Contains(prefix, "/") || strings.Contains(verbBase, "/") {
			continue
		}
		// synthesizer forms (optional)
		if synth != nil {
			for _, form := range synth(verbBase) {
				if form == "" || strings.Contains(form, "ß") {
					continue
				}
				// skip empty / non-lowercase start (Java: Character.isLowerCase)
				r0, _ := utf8.DecodeRuneInString(form)
				if r0 == utf8.RuneError || !unicode.IsLower(r0) {
					continue
				}
				key := prefix + form
				ex.VerbInfos[key] = PrefixInfixVerb{Prefix: prefix, Infix: "", VerbBaseform: verbBase}
			}
		}
		// always: prefix + base infinitive surface
		ex.VerbInfos[prefix+verbBase] = PrefixInfixVerb{Prefix: prefix, Infix: "", VerbBaseform: verbBase}
		// zu + base
		zuKey := prefix + "zu" + verbBase
		ex.VerbInfos[zuKey] = PrefixInfixVerb{Prefix: prefix, Infix: "zu", VerbBaseform: verbBase}
		// VER:EIZ for zu-forms (isSFT unknown without base tags → NON, Java may refine)
		ex.byForm[zuKey] = []tagging.TaggedWord{
			tagging.NewTaggedWord(prefix+verbBase, "VER:EIZ:NON"),
		}
		// Nominalized: Upper(prefix)+base
		nom := tools.UppercaseFirstChar(prefix) + verbBase
		lemma := nom
		ex.byForm[nom] = []tagging.TaggedWord{
			tagging.NewTaggedWord(lemma, "SUB:NOM:SIN:NEU:INF"),
			tagging.NewTaggedWord(lemma, "SUB:AKK:SIN:NEU:INF"),
			tagging.NewTaggedWord(lemma, "SUB:DAT:SIN:NEU:INF"),
		}
		// Genitive + s
		gen := nom + "s"
		ex.byForm[gen] = []tagging.TaggedWord{
			tagging.NewTaggedWord(lemma, "SUB:GEN:SIN:NEU:INF"),
		}
	}
	return ex, sc.Err()
}

// LookupVerb returns PrefixInfixVerb info for a surface (Java verbInfos.get).
func (e *SpellingVerbExpansion) LookupVerb(word string) (PrefixInfixVerb, bool) {
	if e == nil {
		return PrefixInfixVerb{}, false
	}
	v, ok := e.VerbInfos[word]
	return v, ok
}

// Tag implements tagging.WordTagger for fixed nominalized/zu readings.
func (e *SpellingVerbExpansion) Tag(word string) []tagging.TaggedWord {
	if e == nil {
		return nil
	}
	return append([]tagging.TaggedWord(nil), e.byForm[word]...)
}

// Size returns VerbInfos entry count.
func (e *SpellingVerbExpansion) Size() int {
	if e == nil {
		return 0
	}
	return len(e.VerbInfos)
}

var _ tagging.WordTagger = (*SpellingVerbExpansion)(nil)

// tagsForWeise ports GermanTagger.tagsForWeise (idealerweise etc.).
var tagsForWeise = []string{
	"ADJ:AKK:PLU:FEM:GRU:SOL",
	"ADJ:AKK:PLU:MAS:GRU:SOL",
	"ADJ:AKK:PLU:NEU:GRU:SOL",
	"ADJ:AKK:SIN:FEM:GRU:DEF",
	"ADJ:AKK:SIN:FEM:GRU:IND",
	"ADJ:AKK:SIN:FEM:GRU:SOL",
	"ADJ:AKK:SIN:NEU:GRU:DEF",
	"ADJ:NOM:PLU:FEM:GRU:SOL",
	"ADJ:NOM:PLU:MAS:GRU:SOL",
	"ADJ:NOM:PLU:NEU:GRU:SOL",
	"ADJ:NOM:SIN:FEM:GRU:DEF",
	"ADJ:NOM:SIN:FEM:GRU:IND",
	"ADJ:NOM:SIN:FEM:GRU:SOL",
	"ADJ:NOM:SIN:MAS:GRU:DEF",
	"ADJ:NOM:SIN:NEU:GRU:DEF",
	"ADJ:PRD:GRU",
}
