package de

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// GermanCompoundTokenizer ports org.languagetool.tokenizers.de.GermanCompoundTokenizer
// using a faithful jWordSplitter GermanWordSplitter (hideInterfixCharacters=false)
// plus LanguageTool ExtendedGermanWordSplitter.extendedList and constructor addException.
type GermanCompoundTokenizer struct {
	splitter *germanWordSplitter

	// Words is the mutable compound-part lexicon (same map the splitter uses).
	// Exposed for callers that AddWord / LoadHunspellDic after construction.
	Words map[string]struct{}
	// Exceptions mirrors addException surfaces (lowercase) → parts for tests/introspection.
	Exceptions map[string][]string
	// Strict is the constructor strictMode flag.
	Strict bool
	// MinPartLen is setMinimumWordLength (Java default 3, constructor sets 3).
	MinPartLen int
}

var (
	strictInstanceOnce    sync.Once
	nonStrictInstanceOnce sync.Once
	strictInstance        *GermanCompoundTokenizer
	nonStrictInstance     *GermanCompoundTokenizer
)

// NewGermanCompoundTokenizer ports GermanCompoundTokenizer(boolean strictMode).
func NewGermanCompoundTokenizer(strict bool) *GermanCompoundTokenizer {
	// ExtendedGermanWordSplitter(false): hideInterfix=false, words=extendedList()
	words := extendedGermanWordList()
	sp := newGermanWordSplitter(false, words)
	sp.setStrictMode(strict)
	sp.setMinimumWordLength(3)

	t := &GermanCompoundTokenizer{
		splitter:   sp,
		Words:      words,
		Exceptions: map[string][]string{},
		Strict:     strict,
		MinPartLen: 3,
	}
	t.applyConstructorExceptions()
	return t
}

// GetStrictInstance ports GermanCompoundTokenizer.getStrictInstance().
func GetStrictInstance() *GermanCompoundTokenizer {
	strictInstanceOnce.Do(func() {
		strictInstance = NewGermanCompoundTokenizer(true)
	})
	return strictInstance
}

// GetNonStrictInstance ports GermanCompoundTokenizer.getNonStrictInstance().
func GetNonStrictInstance() *GermanCompoundTokenizer {
	nonStrictInstanceOnce.Do(func() {
		nonStrictInstance = NewGermanCompoundTokenizer(false)
	})
	return nonStrictInstance
}

// extendedGermanWordList ports ExtendedGermanWordSplitter.extendedList().
func extendedGermanWordList() map[string]struct{} {
	words := copyWordSet(embeddedGermanDictionaryWords())
	for _, w := range compoundTokenizerExtendedParts {
		words[javaStringToLower(w)] = struct{}{}
	}
	return words
}

// applyConstructorExceptions ports the addException calls in the Java constructor.
func (t *GermanCompoundTokenizer) applyConstructorExceptions() {
	add := func(surface string, parts ...string) {
		t.AddException(surface, parts)
	}
	// Commented in Java (not active):
	// wordSplitter.addException("Maskerade", Collections.singletonList("Maskerade"));
	// wordSplitter.addException("Sportshorts", asList("Sport", "shorts"));
	add("Absolventen", "Absolventen")
	add("Acetat", "Acetat")
	add("Alkoholabstinenz", "Alkohol", "abstinenz")
	add("Androgen", "Androgen")
	add("Auberginen", "Auberginen")
	add("Auckland", "Auckland")
	add("Boston", "Boston")
	add("Brandenburg", "Brandenburg")
	add("Broadcast", "Broadcast")
	add("Buchsbaum", "Buchsbaum")
	add("Chiemsee", "Chiemsee")
	add("Coffein", "Coffein")
	add("Drohnen", "Drohnen")
	add("Eiben", "Eiben")
	add("Eingroschen", "Eingroschen")
	add("Einkomponenten", "Einkomponenten")
	add("Elster", "Elster")
	add("Engineering", "Engineering")
	add("Factoring", "Factoring")
	add("Flexodruck", "Flexo", "druck")
	add("Graviton", "Graviton")
	add("Göttinnen", "Göttinnen")
	add("Hallesche", "Hallesche")
	add("Hinspiel", "Hinspiel")
	add("Homogen", "Homogen")
	add("Kolleggen", "Kolleggen")
	add("Karstadt", "Karstadt")
	add("Kartier", "Kartier")
	add("Kaukasus", "Kaukasus")
	add("Knoblauch", "Knoblauch")
	add("Kollagen", "Kollagen")
	add("Kommerz", "Kommerz")
	add("Mentoring", "Mentoring")
	add("Monarchen", "Monarchen")
	add("Oligarchen", "Oligarchen")
	add("Optimal", "Optimal")
	add("Saunieren", "Saunieren")
	add("Schiessen", "Schiessen")
	add("Spielgeleier", "Spielgeleier")
	add("Halleschen", "Halleschen")
	add("Reinigungstab", "Reinigungs", "tab")
	add("Reinigungstabs", "Reinigungs", "tabs")
	add("Tauschwerte", "Tausch", "werte")
	add("Tauschwertes", "Tausch", "wertes")
	add("Kinderspielen", "Kinder", "spielen")
	add("Buchhaltungstrick", "Buchhaltungs", "trick")
	add("Buchhaltungstricks", "Buchhaltungs", "tricks")
	add("Haushaltstrick", "Haushalts", "trick")
	add("Haushaltstricks", "Haushalts", "tricks")
	add("Verkaufstrick", "Verkaufs", "trick")
	add("Verkaufstricks", "Verkaufs", "tricks")
	add("Ablenkungstrick", "Ablenkungs", "trick")
	add("Ablenkungstricks", "Ablenkungs", "tricks")
	add("Manipulationstrick", "Manipulations", "trick")
	add("Manipulationstricks", "Manipulations", "tricks")
	add("Erziehungstrick", "Erziehungs", "trick")
	add("Erziehungstricks", "Erziehungs", "tricks")
	add("Messetage", "Messe", "tage")
	add("Messetagen", "Messe", "tagen")
	add("karamelligen", "karamelligen") // != Karamel+Ligen
	add("Häkelnadel", "Häkel", "nadel")
	add("Häkelnadeln", "Häkel", "nadeln")
	add("Freiberg", "Freiberg")
	add("Abtestat", "Abtestat")
	add("Abtestaten", "Abtestaten")
	add("Freibergs", "Freibergs")
	add("Kreuzberg", "Kreuzberg")
	add("Kreuzbergs", "Kreuzbergs")
	add("Digitalisierung", "Digitalisierung")
	add("Abtrocknung", "Abtrocknung")
	add("Erlösung", "Erlösung")
	add("Feuerung", "Feuerung")
	add("Aktivierung", "Aktivierung")
	add("Protokollierung", "Protokollierung")
	add("Budgetierung", "Budgetierung")
	add("Faltung", "Faltung")
	add("Anhäufung", "Anhäufung")
	add("Aufkohlung", "Aufkohlung")
	add("Festigung", "Festigung")
	add("Allerheiligen", "Allerheiligen")
	add("Druckerpressen", "Drucker", "pressen")
	add("Habitat", "Habitat")
	add("Augarten", "Augarten")
	add("Auszeit", "Auszeit")
	add("Bewegtbild", "Bewegt", "bild")
	add("Bigband", "Bigband")
	add("Bisexuelle", "Bisexuelle")
	add("Bisexuellen", "Bisexuellen")
	add("Bunsenbrenner", "Bunsenbrenner") // avoid "Bunsenweisheit"
	add("Carbon", "Carbon")
	add("Carsharing", "Carsharing")
	add("Castor", "Castor")
	add("Catering", "Catering")
	add("Cholesterin", "Cholesterin")
	add("Damast", "Damast")
	add("Dispositiv", "Dispositiv")
	add("Emittent", "Emittent")
	add("Emittenten", "Emittenten")
	add("Express", "Express")
	add("Fairness", "Fairness")
	add("Fiberglas", "Fiberglas") // avoid "Fiberthermometer"
	add("Globus", "Globus")
	add("Göttinnen", "Göttinnen")
	add("Illustration", "Illustration")
	add("Muttertag", "Muttertag")   // "Muttertag" needs infix s
	add("Muttertags", "Muttertags") // "Muttertag" needs infix s
	add("Patriarchen", "Patriarchen")
	add("Phosgen", "Phosgen")
	add("Vatertag", "Vatertag")   // "Vatertag" needs infix s
	add("Vatertags", "Vatertags") // "Vatertag" needs infix s
	add("Vaterland", "Vaterland")
	add("Vaterlands", "Vaterlands")
	add("Wehrmacht", "Wehrmacht")
	add("Wehrmachts", "Wehrmachts")
}

// ApplyLanguageToolExtras is retained for callers; NewGermanCompoundTokenizer already
// applies extendedList + constructor exceptions. Idempotent re-add of extended parts.
func (t *GermanCompoundTokenizer) ApplyLanguageToolExtras() {
	if t == nil {
		return
	}
	for _, w := range compoundTokenizerExtendedParts {
		t.AddWord(w)
	}
	// Re-apply constructor exceptions (same as Java list).
	t.applyConstructorExceptions()
}

// LoadWords loads dictionary lines (one word per line) into the splitter lexicon.
func (t *GermanCompoundTokenizer) LoadWords(r io.Reader) error {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		w := tools.JavaStringTrim(sc.Text())
		if w == "" || strings.HasPrefix(w, "#") {
			continue
		}
		t.AddWord(w)
	}
	return sc.Err()
}

// LoadHunspellDic loads surface forms from a Hunspell .dic (word before '/', skip header count).
func (t *GermanCompoundTokenizer) LoadHunspellDic(r io.Reader) error {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	first := true
	for sc.Scan() {
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if first {
			first = false
			if isAllDigits(line) {
				continue
			}
		}
		if i := strings.IndexByte(line, '/'); i >= 0 {
			line = line[:i]
		}
		if i := strings.IndexByte(line, '\t'); i >= 0 {
			line = line[:i]
		}
		line = tools.JavaStringTrim(line)
		if line == "" {
			continue
		}
		t.AddWord(line)
	}
	return sc.Err()
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// AddWord adds a compound part to the jWordSplitter word list (lowercase key).
func (t *GermanCompoundTokenizer) AddWord(w string) {
	if t == nil {
		return
	}
	w = tools.JavaStringTrim(w)
	if w == "" {
		return
	}
	if t.Words == nil {
		t.Words = map[string]struct{}{}
	}
	lw := javaStringToLower(w)
	t.Words[lw] = struct{}{}
	if t.splitter != nil {
		t.splitter.addWord(w)
		// Keep splitter.words and t.Words the same map when possible.
		t.splitter.words = t.Words
		if t.splitter.disambiguator != nil {
			t.splitter.disambiguator.dictionary = t.Words
		}
	}
}

// AddException ports WordSplitter.addException (fixed split for a surface).
func (t *GermanCompoundTokenizer) AddException(surface string, parts []string) {
	if t == nil || surface == "" || len(parts) == 0 {
		return
	}
	if t.Exceptions == nil {
		t.Exceptions = map[string][]string{}
	}
	cp := append([]string(nil), parts...)
	t.Exceptions[javaStringToLower(surface)] = cp
	if t.splitter != nil {
		// Java: exceptionSplits.addSplit(completeWord.toLowerCase(), wordParts)
		// wordParts keep original casing from asList(...).
		t.splitter.addException(surface, cp)
	}
}

// Tokenize ports GermanCompoundTokenizer.tokenize → wordSplitter.splitWord,
// catching InputTooLongException as Collections.singletonList(word).
func (t *GermanCompoundTokenizer) Tokenize(text string) []string {
	if t == nil || t.splitter == nil {
		if text == "" {
			return nil
		}
		return []string{text}
	}
	// Sync min length if caller mutated MinPartLen.
	if t.MinPartLen > 0 {
		t.splitter.setMinimumWordLength(t.MinPartLen)
	}
	t.splitter.setStrictMode(t.Strict)
	// If caller replaced Words map, rebind.
	if t.Words != nil {
		t.splitter.words = t.Words
		if t.splitter.disambiguator != nil {
			t.splitter.disambiguator.dictionary = t.Words
		}
	}
	parts, err := t.splitter.splitWord(text)
	if err != nil {
		if errors.Is(err, errInputTooLong) {
			return []string{text}
		}
		return []string{text}
	}
	return parts
}

// AllSplits ports jWordSplitter AbstractWordSplitter.getAllSplits.
// On InputTooLongException returns nil (callers treat as no candidates); Java throws.
func (t *GermanCompoundTokenizer) AllSplits(text string) [][]string {
	if t == nil || t.splitter == nil || text == "" {
		return nil
	}
	if t.MinPartLen > 0 {
		t.splitter.setMinimumWordLength(t.MinPartLen)
	}
	t.splitter.setStrictMode(t.Strict)
	if t.Words != nil {
		t.splitter.words = t.Words
		if t.splitter.disambiguator != nil {
			t.splitter.disambiguator.dictionary = t.Words
		}
	}
	all, err := t.splitter.getAllSplits(text)
	if err != nil {
		return nil
	}
	// Drop single-part trivial results if any — Java returns multi-part splits only
	// from the loop (always at least 2 pieces per entry).
	return all
}
