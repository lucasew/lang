package suggestions_ordering

import (
	"fmt"
	"math/rand"
)

// Distance ports DetailedDamerauLevenstheinDistance.Distance.
type Distance struct {
	Inserts    int
	Deletes    int
	Replaces   int
	Transposes int
}

func (d Distance) Value() int {
	return d.Inserts + d.Deletes + d.Replaces + d.Transposes
}

func (d Distance) Insert() Distance {
	return Distance{d.Inserts + 1, d.Deletes, d.Replaces, d.Transposes}
}
func (d Distance) Delete() Distance {
	return Distance{d.Inserts, d.Deletes + 1, d.Replaces, d.Transposes}
}
func (d Distance) Replace() Distance {
	return Distance{d.Inserts, d.Deletes, d.Replaces + 1, d.Transposes}
}
func (d Distance) Transpose() Distance {
	return Distance{d.Inserts, d.Deletes, d.Replaces, d.Transposes + 1}
}

func (d Distance) String() string {
	return fmt.Sprintf("Distance{value=%d inserts=%d deletes=%d replaces=%d transposes=%d}",
		d.Value(), d.Inserts, d.Deletes, d.Replaces, d.Transposes)
}

// EditOperation randomly mutates a string (for test generation).
type EditOperation interface {
	Apply(s string) string
}

type seededOp struct {
	r *rand.Rand
}

func (o *seededOp) rng() *rand.Rand {
	if o.r == nil {
		o.r = rand.New(rand.NewSource(1))
	}
	return o.r
}

// Delete removes a random character.
type DeleteOp struct{ seededOp }

func (o *DeleteOp) Apply(s string) string {
	rs := []rune(s)
	if len(rs) <= 1 {
		return ""
	}
	i := o.rng().Intn(len(rs))
	return string(append(append([]rune{}, rs[:i]...), rs[i+1:]...))
}

// Insert inserts a random lowercase letter.
type InsertOp struct{ seededOp }

func (o *InsertOp) Apply(s string) string {
	rs := []rune(s)
	i := o.rng().Intn(len(rs) + 1)
	c := rune('a' + o.rng().Intn(26))
	out := make([]rune, 0, len(rs)+1)
	out = append(out, rs[:i]...)
	out = append(out, c)
	out = append(out, rs[i:]...)
	return string(out)
}

// Replace substitutes a random character.
type ReplaceOp struct{ seededOp }

func (o *ReplaceOp) Apply(s string) string {
	rs := []rune(s)
	if len(rs) == 0 {
		return ""
	}
	i := o.rng().Intn(len(rs))
	rs[i] = rune('a' + o.rng().Intn(26))
	return string(rs)
}

// Transpose swaps two adjacent characters.
type TransposeOp struct{ seededOp }

func (o *TransposeOp) Apply(s string) string {
	rs := []rune(s)
	if len(rs) <= 1 {
		return ""
	}
	i := o.rng().Intn(len(rs) - 1)
	rs[i], rs[i+1] = rs[i+1], rs[i]
	return string(rs)
}

// Compare ports DetailedDamerauLevenstheinDistance.compare (OSA Damerau with op counts).
func Compare(s1, s2 string) Distance {
	if s1 == s2 {
		return Distance{}
	}
	a, b := []rune(s1), []rune(s2)
	inf := len(a) + len(b)
	da := map[rune]int{}
	for _, r := range a {
		da[r] = 0
	}
	for _, r := range b {
		da[r] = 0
	}
	// H[0..len(a)+1][0..len(b)+1]
	h := make([][]Distance, len(a)+2)
	for i := range h {
		h[i] = make([]Distance, len(b)+2)
	}
	for i := 0; i <= len(a); i++ {
		h[i+1][0] = Distance{Inserts: inf}
		h[i+1][1] = Distance{Inserts: i}
	}
	for j := 0; j <= len(b); j++ {
		h[0][j+1] = Distance{Inserts: inf}
		h[1][j+1] = Distance{Inserts: j}
	}
	for i := 1; i <= len(a); i++ {
		db := 0
		for j := 1; j <= len(b); j++ {
			i1 := da[b[j-1]]
			j1 := db
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
				db = j
			}
			substitution := h[i][j].Value() + cost
			insertion := h[i+1][j].Value() + 1
			deletion := h[i][j+1].Value() + 1
			transpose := h[i1][j1].Value() + (i - i1 - 1) + 1 + (j - j1 - 1)
			minV := substitution
			if insertion < minV {
				minV = insertion
			}
			if deletion < minV {
				minV = deletion
			}
			if transpose < minV {
				minV = transpose
			}
			switch minV {
			case substitution:
				if cost == 1 {
					h[i+1][j+1] = h[i][j].Replace()
				} else {
					h[i+1][j+1] = h[i][j]
				}
			case insertion:
				h[i+1][j+1] = h[i+1][j].Insert()
			case deletion:
				h[i+1][j+1] = h[i][j+1].Delete()
			default: // transpose
				transposeCost := (i - i1 - 1) + 1 + (j - j1 - 1)
				v := h[i1][j1]
				for k := 0; k < transposeCost; k++ {
					v = v.Transpose()
				}
				h[i+1][j+1] = v
			}
		}
		da[a[i-1]] = i
	}
	return h[len(a)+1][len(b)+1]
}
