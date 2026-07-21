package nl

import (
	"bufio"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
)

// CompoundAcceptor ports org.languagetool.rules.nl.CompoundAcceptor.
// Accepts 2-part Dutch compounds the speller rejects.
//
// Tagger/speller probes (isNoun, spellingOk, …) are injectable. When unset:
//   - TagPOS/IsNoun/IsExistingWord/IsGeographical: fail-closed (false)
//   - SpellingOk: uses FilterDict when wired; otherwise fail-closed (false)
//
// KnownWords is not in Java; kept only as an optional full-word test whitelist.
type CompoundAcceptor struct {
	NoS               map[string]struct{}
	NeedsS            map[string]struct{}
	AlwaysNeedsS      map[string]struct{}
	AlwaysNeedsHyphen map[string]struct{}
	Directions        map[string]struct{} // geographicalDirections (case-sensitive as loaded)
	Part1Exceptions   map[string]struct{}
	Part2Exceptions   map[string]struct{}
	AcronymExceptions map[string]struct{} // mixed case as in file
	// KnownWords optional full-word whitelist (tests only; not in Java).
	KnownWords  map[string]struct{}
	MaxWordSize int

	// IsNoun ports isNoun (POS starts with ZNW, not part2Exceptions).
	IsNoun func(word string) bool
	// IsExistingWord ports isExistingWord (any non-null POS).
	IsExistingWord func(word string) bool
	// IsGeographical ports isGeographicalCompound (POS starts with ENM:LOC).
	IsGeographical func(word string) bool
	// SpellingOk ports spellingOk (normalCase + speller clean).
	SpellingOk func(word string) bool
	// TagPOS optional: when set and IsNoun/etc nil, used for POS checks.
	TagPOS func(word string) []string
}

// DefaultCompoundAcceptor is the process singleton (Java CompoundAcceptor.INSTANCE).
// Word lists load once at package init when resources are discoverable.
var DefaultCompoundAcceptor = newDefaultCompoundAcceptor()

func newDefaultCompoundAcceptor() *CompoundAcceptor {
	c := NewCompoundAcceptor()
	_ = c.LoadDefaultWordLists()
	return c
}

// BindDefaultCompoundAcceptorFilters ports Java CompoundAcceptor fields:
//
//	dutchTagger.getPostags → TagPOS (FilterGetPostags when dutch.dict wired)
//	spellingOk speller → FilterDict when nl_NL.dict wired (via spellingOk default)
//
// Call after TryWireDutchFilterTagger / TryWireDutchFilterSpeller.
// Does not invent POS when dicts are missing (hooks stay fail-closed).
func BindDefaultCompoundAcceptorFilters() {
	if DefaultCompoundAcceptor == nil {
		return
	}
	if FilterTaggerAvailable() {
		DefaultCompoundAcceptor.TagPOS = FilterGetPostags
	}
}

func NewCompoundAcceptor() *CompoundAcceptor {
	return &CompoundAcceptor{
		NoS:               map[string]struct{}{},
		NeedsS:            map[string]struct{}{},
		AlwaysNeedsS:      map[string]struct{}{},
		AlwaysNeedsHyphen: map[string]struct{}{},
		Directions:        map[string]struct{}{},
		Part1Exceptions:   map[string]struct{}{},
		Part2Exceptions:   map[string]struct{}{},
		AcronymExceptions: map[string]struct{}{},
		KnownWords:        map[string]struct{}{},
		MaxWordSize:       35, // Java MAX_WORD_SIZE
	}
}

// collidingVowels ports CompoundAcceptor.collidingVowels.
var collidingVowels = map[string]struct{}{
	"aa": {}, "ae": {}, "ai": {}, "au": {},
	"ee": {}, "ée": {}, "ei": {}, "éi": {}, "eu": {}, "éu": {},
	"ie": {}, "ii": {}, "ij": {},
	"oe": {}, "oi": {}, "oo": {}, "ou": {},
	"ui": {}, "uu": {},
}

var (
	// Java acronymPattern = [A-Z]{2,4}-
	acronymPattern = regexp.MustCompile(`^[A-Z]{2,4}-$`)
	// Java specialAcronymPattern = [A-Za-z]{2,4}-
	specialAcronymPattern = regexp.MustCompile(`^[A-Za-z]{2,4}-$`)
	// Java normalCasePattern = [A-Za-z][a-zé]*
	normalCasePattern = regexp.MustCompile(`^[A-Za-z][a-zé]*$`)
)

func loadSetPreserveCase(r io.Reader, dest map[string]struct{}) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		w := strings.TrimSpace(sc.Text())
		if w == "" || strings.HasPrefix(w, "#") {
			continue
		}
		if i := strings.IndexByte(w, '#'); i >= 0 {
			w = strings.TrimSpace(w[:i])
		}
		if w != "" {
			dest[w] = struct{}{}
		}
	}
	return sc.Err()
}

func loadSetFile(path string, dest map[string]struct{}) error {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return loadSetPreserveCase(f, dest)
}

// LoadDefaultWordLists ports CompoundAcceptor() constructor wordListLoader loads.
// Missing files are skipped (partial init).
func (c *CompoundAcceptor) LoadDefaultWordLists() error {
	if c == nil {
		return nil
	}
	pairs := []struct {
		rel  string
		dest map[string]struct{}
	}{
		{"nl/compound_acceptor/no_s.txt", c.NoS},
		{"nl/compound_acceptor/needs_s.txt", c.NeedsS},
		{"nl/compound_acceptor/directions.txt", c.Directions},
		{"nl/compound_acceptor/always_needs_s.txt", c.AlwaysNeedsS},
		{"nl/compound_acceptor/always_needs_hyphen.txt", c.AlwaysNeedsHyphen},
		{"nl/compound_acceptor/part1_exceptions.txt", c.Part1Exceptions},
		{"nl/compound_acceptor/part2_exceptions.txt", c.Part2Exceptions},
		{"nl/compound_acceptor/acronym_exceptions.txt", c.AcronymExceptions},
	}
	for _, p := range pairs {
		path := spelling.DiscoverSpellingResource(p.rel)
		if path == "" {
			continue
		}
		if err := loadSetFile(path, p.dest); err != nil {
			return err
		}
	}
	return nil
}

// Load helpers for tests / custom streams (preserve case like Java loadWords).
func (c *CompoundAcceptor) LoadNoS(r io.Reader) error {
	return loadSetPreserveCase(r, c.NoS)
}
func (c *CompoundAcceptor) LoadNeedsS(r io.Reader) error {
	return loadSetPreserveCase(r, c.NeedsS)
}
func (c *CompoundAcceptor) LoadDirections(r io.Reader) error {
	return loadSetPreserveCase(r, c.Directions)
}
func (c *CompoundAcceptor) LoadAlwaysNeedsS(r io.Reader) error {
	return loadSetPreserveCase(r, c.AlwaysNeedsS)
}
func (c *CompoundAcceptor) LoadAlwaysNeedsHyphen(r io.Reader) error {
	return loadSetPreserveCase(r, c.AlwaysNeedsHyphen)
}
func (c *CompoundAcceptor) LoadPart1Exceptions(r io.Reader) error {
	return loadSetPreserveCase(r, c.Part1Exceptions)
}
func (c *CompoundAcceptor) LoadPart2Exceptions(r io.Reader) error {
	return loadSetPreserveCase(r, c.Part2Exceptions)
}
func (c *CompoundAcceptor) LoadAcronymExceptions(r io.Reader) error {
	return loadSetPreserveCase(r, c.AcronymExceptions)
}

// Accept is the public name used by MorfologikDutchSpellerRule; ports acceptCompound(String).
func (c *CompoundAcceptor) Accept(word string) bool {
	return c.AcceptCompound(word)
}

// AcceptCompound ports acceptCompound(String word).
func (c *CompoundAcceptor) AcceptCompound(word string) bool {
	if c == nil || word == "" {
		return false
	}
	// Java word.length() (UTF-16); for Dutch BMP same as rune count.
	if tokenizers.UTF16Len(word) > c.MaxWordSize {
		return false
	}
	if _, ok := c.KnownWords[strings.ToLower(word)]; ok {
		return true
	}
	runes := []rune(word)
	// Java: for (i = 3; i < word.length() - 3; i++)
	for i := 3; i < len(runes)-3; i++ {
		part1 := string(runes[:i])
		part2 := string(runes[i:])
		if part1 != part2 && c.AcceptCompoundParts(part1, part2) {
			return true
		}
	}
	return false
}

// GetParts ports getParts: first accepting split or empty.
func (c *CompoundAcceptor) GetParts(word string) []string {
	if c == nil || word == "" {
		return nil
	}
	if tokenizers.UTF16Len(word) > c.MaxWordSize {
		return nil
	}
	runes := []rune(word)
	for i := 3; i < len(runes)-3; i++ {
		part1 := string(runes[:i])
		part2 := string(runes[i:])
		if part1 != part2 && c.AcceptCompoundParts(part1, part2) {
			return []string{part1, part2}
		}
	}
	return nil
}

// AcceptCompoundParts ports acceptCompound(String part1, String part2).
func (c *CompoundAcceptor) AcceptCompoundParts(part1, part2 string) bool {
	if c == nil || part1 == "" || part2 == "" {
		return false
	}
	part1lc := strings.ToLower(part1)

	// branch: part1 ends with "s" and not exception/alwaysNeedsS/noS/hyphen form
	if strings.HasSuffix(part1, "s") &&
		!c.has(c.Part1Exceptions, part1[:len(part1)-1]) &&
		!c.has(c.AlwaysNeedsS, part1) &&
		!c.has(c.NoS, part1) &&
		!strings.Contains(part1, "-") {
		for suffix := range c.AlwaysNeedsS {
			if strings.HasSuffix(part1lc, suffix) {
				// Java: isNoun(part2) && isExistingWord(part1 without s) && spellingOk(part2)
				return c.isNoun(part2) &&
					c.isExistingWord(part1lc[:len(part1lc)-1]) &&
					c.spellingOk(part2)
			}
		}
		// Java: needsS.contains(part1lc) && isNoun(part2) && spellingOk(part1 without s) && spellingOk(part2)
		return c.has(c.NeedsS, part1lc) &&
			c.isNoun(part2) &&
			c.spellingOk(part1[:len(part1)-1]) &&
			c.spellingOk(part2)
	}

	if c.has(c.Directions, part1) {
		return c.isGeographical(part2)
	}

	if strings.HasSuffix(part1, "-") {
		// abbreviations / always-hyphen prefixes
		return (c.acronymOk(part1) || c.has(c.AlwaysNeedsHyphen, part1lc)) && c.spellingOk(part2)
	}

	if strings.HasPrefix(part2, "-") {
		// vowel collision compounds (politie-eenheid)
		p2 := part2[1:]
		return c.has(c.NoS, part1lc) &&
			c.isNoun(p2) &&
			c.spellingOk(part1) &&
			c.spellingOk(p2) &&
			c.hasCollidingVowels(part1, p2)
	}

	// default no-s compound
	return (c.has(c.NoS, part1lc) || c.has(c.Part1Exceptions, part1lc)) &&
		c.isNoun(part2) &&
		c.spellingOk(part1) &&
		!c.hasCollidingVowels(part1, part2)
}

func (c *CompoundAcceptor) has(m map[string]struct{}, key string) bool {
	if c == nil || m == nil || key == "" {
		return false
	}
	_, ok := m[key]
	return ok
}

func (c *CompoundAcceptor) isNoun(word string) bool {
	if c == nil || word == "" {
		return false
	}
	if c.has(c.Part2Exceptions, word) {
		return false
	}
	if c.IsNoun != nil {
		return c.IsNoun(word)
	}
	if c.TagPOS != nil {
		for _, t := range c.TagPOS(word) {
			if strings.HasPrefix(t, "ZNW") {
				return true
			}
		}
	}
	return false
}

func (c *CompoundAcceptor) isExistingWord(word string) bool {
	if c == nil || word == "" {
		return false
	}
	if c.IsExistingWord != nil {
		return c.IsExistingWord(word)
	}
	if c.TagPOS != nil {
		for _, t := range c.TagPOS(word) {
			if t != "" {
				return true
			}
		}
	}
	return false
}

func (c *CompoundAcceptor) isGeographical(word string) bool {
	if c == nil || word == "" {
		return false
	}
	if c.IsGeographical != nil {
		return c.IsGeographical(word)
	}
	if c.TagPOS != nil {
		for _, t := range c.TagPOS(word) {
			if strings.HasPrefix(t, "ENM:LOC") {
				return true
			}
		}
	}
	return false
}

func (c *CompoundAcceptor) spellingOk(word string) bool {
	if c == nil || word == "" {
		return false
	}
	if c.SpellingOk != nil {
		return c.SpellingOk(word)
	}
	// Java normalCasePattern first
	if !normalCasePattern.MatchString(word) {
		return false
	}
	// Fail-closed without a wired Dutch speller dict (do not invent accept).
	if !FilterDictAvailable() {
		return false
	}
	return !FilterDictIsMisspelled(word)
}

func (c *CompoundAcceptor) hasCollidingVowels(part1, part2 string) bool {
	if part1 == "" || part2 == "" {
		return false
	}
	r1, _ := utf8.DecodeLastRuneInString(part1)
	r2, _ := utf8.DecodeRuneInString(part2)
	if r1 == utf8.RuneError || r2 == utf8.RuneError {
		return false
	}
	pair := strings.ToLower(string(r1) + string(r2))
	_, ok := collidingVowels[pair]
	return ok
}

// acronymOk ports private acronymOk.
func (c *CompoundAcceptor) acronymOk(nonCompound string) bool {
	if c == nil {
		return false
	}
	if acronymPattern.MatchString(nonCompound) {
		// IRA- style: accepted unless exception upper equals stem
		stem := nonCompound[:len(nonCompound)-1]
		for ex := range c.AcronymExceptions {
			if strings.ToUpper(ex) == stem {
				return false
			}
		}
		return true
	}
	if specialAcronymPattern.MatchString(nonCompound) {
		// special casing: must be listed exactly without trailing '-'
		stem := nonCompound[:len(nonCompound)-1]
		return c.has(c.AcronymExceptions, stem)
	}
	return false
}
