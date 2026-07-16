package nl

import (
	"bufio"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// CompoundAcceptor ports rules.nl.CompoundAcceptor for 2-part compounds.
type CompoundAcceptor struct {
	NoS              map[string]struct{}
	NeedsS           map[string]struct{}
	AlwaysNeedsS     map[string]struct{}
	AlwaysNeedsHyphen map[string]struct{}
	// KnownWords optional full-word whitelist.
	KnownWords map[string]struct{}
	MaxWordSize int
}

// DefaultCompoundAcceptor is the process singleton.
var DefaultCompoundAcceptor = NewCompoundAcceptor()

func NewCompoundAcceptor() *CompoundAcceptor {
	return &CompoundAcceptor{
		NoS:               map[string]struct{}{},
		NeedsS:            map[string]struct{}{},
		AlwaysNeedsS:      map[string]struct{}{},
		AlwaysNeedsHyphen: map[string]struct{}{},
		KnownWords:        map[string]struct{}{},
		MaxWordSize:       35,
	}
}

func loadSet(r io.Reader, dest map[string]struct{}) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		w := strings.TrimSpace(sc.Text())
		if w == "" || strings.HasPrefix(w, "#") {
			continue
		}
		dest[strings.ToLower(w)] = struct{}{}
	}
	return sc.Err()
}

func (c *CompoundAcceptor) LoadNoS(r io.Reader) error   { return loadSet(r, c.NoS) }
func (c *CompoundAcceptor) LoadNeedsS(r io.Reader) error { return loadSet(r, c.NeedsS) }

// Accept reports whether word is an acceptable 2-part compound.
func (c *CompoundAcceptor) Accept(word string) bool {
	if c == nil || word == "" {
		return false
	}
	if utf8.RuneCountInString(word) > c.MaxWordSize {
		return false
	}
	if _, ok := c.KnownWords[strings.ToLower(word)]; ok {
		return true
	}
	// hyphenated acronyms like "TV-serie"
	if i := strings.IndexByte(word, '-'); i > 0 {
		left, right := word[:i], word[i+1:]
		if isAcronym(left) && len(right) >= 2 {
			return true
		}
		if _, ok := c.AlwaysNeedsHyphen[strings.ToLower(left)]; ok {
			return true
		}
	}
	// try split into two lowercase parts
	lw := strings.ToLower(word)
	for i := 3; i < len(lw)-2; i++ {
		p1, p2 := lw[:i], lw[i:]
		if c.acceptParts(p1, p2, false) {
			return true
		}
		// with connecting -s-
		if i >= 4 && lw[i-1] == 's' {
			if c.acceptParts(lw[:i-1], p2, true) {
				return true
			}
		}
	}
	return false
}

func (c *CompoundAcceptor) acceptParts(p1, p2 string, withS bool) bool {
	if _, ok := c.NoS[p1]; ok && withS {
		return false
	}
	if _, ok := c.NeedsS[p1]; ok && !withS {
		return false
	}
	if _, ok := c.AlwaysNeedsS[p1]; ok && !withS {
		return false
	}
	// both parts must be known if lists populated; otherwise accept if either list has p1
	if len(c.NoS) == 0 && len(c.NeedsS) == 0 && len(c.KnownWords) == 0 {
		return false
	}
	if len(c.KnownWords) > 0 {
		_, k1 := c.KnownWords[p1]
		_, k2 := c.KnownWords[p2]
		return k1 && k2
	}
	_, inNo := c.NoS[p1]
	_, inNeeds := c.NeedsS[p1]
	return (inNo || inNeeds) && len(p2) >= 3
}

func isAcronym(s string) bool {
	if len(s) < 2 || len(s) > 4 {
		return false
	}
	for _, r := range s {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
