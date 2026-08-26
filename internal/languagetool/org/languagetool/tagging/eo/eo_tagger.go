package eo

// EsperantoTagger ports org.languagetool.tagging.eo.EsperantoTagger (rule-based Tagger,
// not BaseTagger/Morfologik). Closed-class words come from manual-tagger.txt; open-class
// morphology uses endings + verb-tr/verb-ntr/root-ant-at resource lists.

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Resource paths as Java JLanguageTool.getDataBroker() loads them.
const (
	// ManualTaggerResourcePath is resource-dir path: /eo/manual-tagger.txt
	ManualTaggerResourcePath = "/eo/manual-tagger.txt"
	// VerbTrRulesPath is rules-dir path: /eo/verb-tr.txt
	VerbTrRulesPath = "/eo/verb-tr.txt"
	// VerbNtrRulesPath is rules-dir path: /eo/verb-ntr.txt
	VerbNtrRulesPath = "/eo/verb-ntr.txt"
	// RootAntAtRulesPath is rules-dir path: /eo/root-ant-at.txt
	RootAntAtRulesPath = "/eo/root-ant-at.txt"
)

// Verbs always end with this pattern.
var (
	patternVerb   = regexp.MustCompile("(..+)(as|os|is|us|u|i)$")
	patternPrefix = regexp.MustCompile("^(?:mal|mis|ek|re|fi|ne)(.*)")
	patternSuffix = regexp.MustCompile("(.*)(?:ad|aĉ|eg|et)i$")

	// Participles -ant-, -int-, -ont-, -at-, -it-, -ot-
	// Groups: 1=stem+...t, 2=root, 3=[aio], 4=n?, 5=[aoe], 6=j?, 7=n?
	patternParticiple = regexp.MustCompile("((..+)([aio])(n?)t)([aoe])(j?)(n?)$")

	// Pattern 'tabelvortoj'.
	// Groups: 1=prefix, 2=[uoae], 3=j?, 4=n?, 5=am|al|es|el|om
	patternTabelvorto = regexp.MustCompile("^(i|ti|ki|ĉi|neni)(?:(?:([uoae])(j?)(n?))|(am|al|es|el|om))$")

	// Pattern of 'tabelvortoj' which are also tagged adverbs.
	patternTabelvortoAdverb = regexp.MustCompile("^(?:ti|i|ĉi|neni)(?:am|om|el|e)$")
)

// EsperantoTagger ports org.languagetool.tagging.eo.EsperantoTagger.
type EsperantoTagger struct {
	mu sync.Mutex

	// manualTagger tags the closed Esperanto word list (no regular ending).
	manualTagger *tagging.ManualTagger

	setTransitiveVerbs   map[string]struct{}
	setIntransitiveVerbs map[string]struct{}
	setNonParticiple     map[string]struct{}

	initErr error
	inited  bool
}

// NewEsperantoTagger ports the Java no-arg constructor (lazy resource load on first Tag).
func NewEsperantoTagger() *EsperantoTagger {
	return &EsperantoTagger{}
}

// xSystemToUnicode ports EsperantoTagger.xSystemToUnicode: "jxauxdo" → "ĵaŭdo".
// Invoked only on already-lowercased forms. Iterates by rune (Java charAt on BMP).
func xSystemToUnicode(s string) string {
	runes := []rune(s)
	var b strings.Builder
	b.Grow(len(s))
	length := len(runes)
	for i := 0; i < length; i++ {
		c1 := runes[i]
		c2 := rune(' ')
		if i+1 < length {
			c2 = runes[i+1]
		}
		if c2 == 'x' {
			switch c1 {
			case 'c':
				b.WriteRune('ĉ')
				i++
			case 'g':
				b.WriteRune('ĝ')
				i++
			case 'h':
				b.WriteRune('ĥ')
				i++
			case 'j':
				b.WriteRune('ĵ')
				i++
			case 's':
				b.WriteRune('ŝ')
				i++
			case 'u':
				b.WriteRune('ŭ')
				i++
			default:
				b.WriteRune(c1)
			}
		} else {
			b.WriteRune(c1)
		}
	}
	return b.String()
}

// loadWords ports EsperantoTagger.loadWords: UTF-8, one word per line, skip #/empty.
func loadWords(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	words := make(map[string]struct{})
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		words[line] = struct{}{}
	}
	return words, sc.Err()
}

// lazyInit ports EsperantoTagger.lazyInit (synchronized).
func (t *EsperantoTagger) lazyInit() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.inited {
		return t.initErr
	}
	t.inited = true

	manualPath := DiscoverEOManualTagger()
	if manualPath == "" {
		t.initErr = errEOResourcesMissing("manual-tagger.txt")
		return t.initErr
	}
	f, err := os.Open(manualPath)
	if err != nil {
		t.initErr = err
		return t.initErr
	}
	mt, err := tagging.NewManualTagger(f)
	f.Close()
	if err != nil {
		t.initErr = err
		return t.initErr
	}
	t.manualTagger = mt

	trPath := DiscoverEORulesFile("verb-tr.txt")
	ntrPath := DiscoverEORulesFile("verb-ntr.txt")
	rootPath := DiscoverEORulesFile("root-ant-at.txt")
	if trPath == "" || ntrPath == "" || rootPath == "" {
		t.initErr = errEOResourcesMissing("verb-tr.txt / verb-ntr.txt / root-ant-at.txt")
		return t.initErr
	}
	if t.setTransitiveVerbs, err = loadWords(trPath); err != nil {
		t.initErr = err
		return t.initErr
	}
	if t.setIntransitiveVerbs, err = loadWords(ntrPath); err != nil {
		t.initErr = err
		return t.initErr
	}
	if t.setNonParticiple, err = loadWords(rootPath); err != nil {
		t.initErr = err
		return t.initErr
	}
	return nil
}

type eoResError string

func (e eoResError) Error() string { return string(e) }

func errEOResourcesMissing(what string) error {
	return eoResError("esperanto tagger resources missing: " + what +
		" (set LANG_ESPERANTO_RESOURCE_DIR / LANG_ESPERANTO_RULES_DIR or use inspiration submodule)")
}

// findTransitivity ports EsperantoTagger.findTransitivity.
// Returns "tr", "nt", "tn", or "xx".
func (t *EsperantoTagger) findTransitivity(verb string) string {
	if strings.HasSuffix(verb, "iĝi") {
		return "nt"
	} else if strings.HasSuffix(verb, "igi") {
		// The verb "memmortigi" is strange: even though it ends in -igi, it is intransitive.
		if verb == "memmortigi" {
			return "nt"
		}
		return "tr"
	}

	// This loop executes only once for most verbs (or very few times).
	for {
		_, isTransitive := t.setTransitiveVerbs[verb]
		_, isIntransitive := t.setIntransitiveVerbs[verb]

		if isTransitive {
			if isIntransitive {
				return "tn"
			}
			return "tr"
		} else if isIntransitive {
			return "nt"
		}

		// Strip mal-/mis-/ek-/re-/fi-/ne- or -ad/-aĉ/-eg/-et and retry.
		if m := patternPrefix.FindStringSubmatch(verb); m != nil {
			verb = m[1]
			continue
		}
		if m := patternSuffix.FindStringSubmatch(verb); m != nil {
			verb = m[1] + "i"
			continue
		}
		break
	}
	return "xx"
}

// Tag ports EsperantoTagger.tag.
func (t *EsperantoTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	if err := t.lazyInit(); err != nil {
		// Fail closed: null POS for every token (no invent tags).
		out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
		for _, word := range sentenceTokens {
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(
				[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}, 0))
		}
		return out
	}

	tokenReadings := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	for _, word := range sentenceTokens {
		var l []*languagetool.AnalyzedToken

		// Java: word.length() is UTF-16 code units.
		wLen := tagging.UTF16Len(word)
		if wLen > 50 {
			// Avoid excessively long computation times for long (probably artificial) tokens.
			l = append(l, languagetool.NewAnalyzedToken(word, nil, nil))
		} else if wLen > 1 {
			// No Esperanto word is made of one letter only.
			// Lemma uses lower case + Unicode (x-system converted).
			lWord := xSystemToUnicode(strings.ToLower(word))
			manualTags := t.manualTagger.Tag(lWord)

			if len(manualTags) > 0 {
				// Closed word: known lemmas and tags.
				for _, manualTag := range manualTags {
					l = append(l, toAnalyzedToken(word, manualTag.PosTag, manualTag.Lemma, false, false))
				}
			} else {
				// Open word: ending / verb dictionary.

				// Tiu, kiu (tabelvortoj).
				if m := patternTabelvorto.FindStringSubmatch(lWord); m != nil {
					type1Group := strings.ToLower(firstRune(m[1]))
					type2Group := m[2]
					plGroup := m[3]
					accGroup := m[4]
					type3Group := m[5]

					var accusative, plural, typ string
					// Java null groups for the non-taken alternative: in Go, empty
					// type2 means the (am|al|es|el|om) branch matched.
					if type2Group == "" {
						accusative = "xxx"
						plural = " pn "
						typ = strings.ToLower(type3Group)
					} else {
						if strings.EqualFold(accGroup, "n") {
							accusative = "akz"
						} else {
							accusative = "nak"
						}
						if strings.EqualFold(plGroup, "j") {
							plural = " pl "
						} else {
							plural = " np "
						}
						typ = strings.ToLower(type2Group)
					}

					pos := "T " + accusative + plural + type1Group + " " + typ
					l = append(l, toAnalyzedToken(word, pos, "", false, true))

					if patternTabelvortoAdverb.MatchString(lWord) {
						l = append(l, toAnalyzedToken(word, "E nak", lWord, false, false))
					}

				} else if strings.HasSuffix(lWord, "o") {
					l = append(l, toAnalyzedToken(word, "O nak np", lWord, false, false))
				} else if utf8.RuneCountInString(lWord) >= 2 && strings.HasSuffix(lWord, "'") {
					// Java: lWord.length() >= 2 && endsWith("'")
					lemma := lWord[:len(lWord)-1] + "o"
					l = append(l, toAnalyzedToken(word, "O nak np", lemma, false, false))
				} else if strings.HasSuffix(lWord, "oj") {
					l = append(l, toAnalyzedToken(word, "O nak pl", lWord[:len(lWord)-1], false, false))
				} else if strings.HasSuffix(lWord, "on") {
					l = append(l, toAnalyzedToken(word, "O akz np", lWord[:len(lWord)-1], false, false))
				} else if strings.HasSuffix(lWord, "ojn") {
					l = append(l, toAnalyzedToken(word, "O akz pl", lWord[:len(lWord)-2], false, false))

				} else if strings.HasSuffix(lWord, "a") {
					l = append(l, toAnalyzedToken(word, "A nak np", lWord, false, false))
				} else if strings.HasSuffix(lWord, "aj") {
					l = append(l, toAnalyzedToken(word, "A nak pl", lWord[:len(lWord)-1], false, false))
				} else if strings.HasSuffix(lWord, "an") {
					l = append(l, toAnalyzedToken(word, "A akz np", lWord[:len(lWord)-1], false, false))
				} else if strings.HasSuffix(lWord, "ajn") {
					l = append(l, toAnalyzedToken(word, "A akz pl", lWord[:len(lWord)-2], false, false))

				} else if strings.HasSuffix(lWord, "e") {
					l = append(l, toAnalyzedToken(word, "E nak", lWord, false, false))
				} else if strings.HasSuffix(lWord, "en") {
					l = append(l, toAnalyzedToken(word, "E akz", lWord[:len(lWord)-1], false, false))

				} else if m := patternVerb.FindStringSubmatch(lWord); m != nil {
					verb := m[1] + "i"
					tense := m[2]
					transitive := t.findTransitivity(verb)
					pos := "V " + transitive + " " + tense
					l = append(l, toAnalyzedToken(word, pos, verb, false, false))

				} else {
					// Irregular word (no tag).
					l = append(l, languagetool.NewAnalyzedToken(word, nil, nil))
				}

				// Participle (can be combined with other tags).
				if m := patternParticiple.FindStringSubmatch(lWord); m != nil {
					if _, skip := t.setNonParticiple[m[1]]; !skip {
						verb := m[2] + "i"
						aio := m[3]
						antAt := "-"
						if m[4] == "n" {
							antAt = "n"
						}
						aoe := m[5]
						plural := "np"
						if m[6] == "j" {
							plural = "pl"
						}
						accusative := "nak"
						if m[7] == "n" {
							accusative = "akz"
						}
						transitive := t.findTransitivity(verb)
						pos := "C " + accusative + " " + plural + " " +
							transitive + " " + aio + " " + antAt + " " + aoe
						l = append(l, toAnalyzedToken(word, pos, verb, false, false))
					}
				}
			}
		} else {
			// Single letter word (no tag).
			l = append(l, languagetool.NewAnalyzedToken(word, nil, nil))
		}
		// Java: new AnalyzedTokenReadings(l, 0)
		tokenReadings = append(tokenReadings, languagetool.NewAnalyzedTokenReadingsList(l, 0))
	}
	return tokenReadings
}

// CreateNullToken ports EsperantoTagger.createNullToken.
func (t *EsperantoTagger) CreateNullToken(token string, startPos int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, nil, nil), startPos)
}

// CreateToken ports EsperantoTagger.createToken (lemma null).
func (t *EsperantoTagger) CreateToken(token, posTag string) *languagetool.AnalyzedToken {
	return toAnalyzedToken(token, posTag, "", false, true)
}

func toAnalyzedToken(surface, pos, lemma string, posNull, lemmaNull bool) *languagetool.AnalyzedToken {
	var p, l *string
	if !posNull {
		p = &pos
	}
	if !lemmaNull {
		l = &lemma
	}
	return languagetool.NewAnalyzedToken(surface, p, l)
}

// firstRune returns the first Unicode code point as a string (Java substring(0,1) on BMP).
func firstRune(s string) string {
	if s == "" {
		return ""
	}
	r, _ := utf8.DecodeRuneInString(s)
	return string(r)
}

// ---------------------------------------------------------------------------
// Resource discovery (same official files as Java under inspiration/...)
// ---------------------------------------------------------------------------

// DiscoverEOManualTagger finds resource/eo/manual-tagger.txt.
// Order: LANG_ESPERANTO_MANUAL_TAGGER, LANG_ESPERANTO_RESOURCE_DIR/manual-tagger.txt,
// walk-up inspiration module path.
func DiscoverEOManualTagger() string {
	if p := os.Getenv("LANG_ESPERANTO_MANUAL_TAGGER"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if dir := os.Getenv("LANG_ESPERANTO_RESOURCE_DIR"); dir != "" {
		p := filepath.Join(dir, "manual-tagger.txt")
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	return walkUpEOFile(filepath.Join("resource", "eo", "manual-tagger.txt"))
}

// DiscoverEORulesFile finds rules/eo/<name> (verb-tr.txt, verb-ntr.txt, root-ant-at.txt).
// Order: LANG_ESPERANTO_RULES_DIR/<name>, walk-up inspiration module path.
func DiscoverEORulesFile(name string) string {
	if dir := os.Getenv("LANG_ESPERANTO_RULES_DIR"); dir != "" {
		p := filepath.Join(dir, name)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	return walkUpEOFile(filepath.Join("rules", "eo", name))
}

// EOResourcesAvailable reports whether all four official resource files resolve.
func EOResourcesAvailable() bool {
	return DiscoverEOManualTagger() != "" &&
		DiscoverEORulesFile("verb-tr.txt") != "" &&
		DiscoverEORulesFile("verb-ntr.txt") != "" &&
		DiscoverEORulesFile("root-ant-at.txt") != ""
}

func walkUpEOFile(relUnderOrgLT string) string {
	// inspiration/.../eo/src/main/resources/org/languagetool/<resource|rules>/eo/...
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "eo",
		"src", "main", "resources", "org", "languagetool", relUnderOrgLT)
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
