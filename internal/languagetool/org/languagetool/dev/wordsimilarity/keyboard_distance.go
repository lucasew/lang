package wordsimilarity

import "unicode"

// KeyboardDistance ports org.languagetool.dev.wordsimilarity.KeyboardDistance.
type KeyboardDistance interface {
	GetDistance(c1, c2 rune) float32
}

// BaseKeyboardDistance ports BaseKeyboardDistance (Manhattan distance on a key layout).
type BaseKeyboardDistance struct {
	Keys [][]rune
}

// GetDistance returns Manhattan distance between keys (case-insensitive).
func (b *BaseKeyboardDistance) GetDistance(c1, c2 rune) float32 {
	p1 := b.position(c1)
	p2 := b.position(c2)
	dr := p1.row - p2.row
	if dr < 0 {
		dr = -dr
	}
	dc := p1.col - p2.col
	if dc < 0 {
		dc = -dc
	}
	return float32(dr + dc)
}

type keyPos struct{ row, col int }

func (b *BaseKeyboardDistance) position(search rune) keyPos {
	lower := unicode.ToLower(search)
	for r, row := range b.Keys {
		for c, ch := range row {
			if ch == lower {
				return keyPos{row: r, col: c}
			}
		}
	}
	panic("Could not find key on keyboard - only letters are supported")
}

// GermanQwertzKeyboardDistance ports German QWERTZ layout distances.
type GermanQwertzKeyboardDistance struct {
	BaseKeyboardDistance
}

// NewGermanQwertzKeyboardDistance returns a German QWERTZ distance calculator.
func NewGermanQwertzKeyboardDistance() *GermanQwertzKeyboardDistance {
	keys := [][]rune{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 'ß'},
		{'q', 'w', 'e', 'r', 't', 'z', 'u', 'i', 'o', 'p', 'ü'},
		{'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'ö', 'ä'},
		{'y', 'x', 'c', 'v', 'b', 'n', 'm'},
	}
	return &GermanQwertzKeyboardDistance{BaseKeyboardDistance{Keys: keys}}
}
