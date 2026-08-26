package suggestions_ordering

// Faithful port of
// org.languagetool.rules.spelling.morfologik.suggestions_ordering.DetailedDamerauLevenstheinDistance
// (Java String length/charAt = UTF-16 code units).

import (
	"fmt"
	"math/rand"
	"time"
	"unicode/utf16"
)

// Distance ports DetailedDamerauLevenstheinDistance.Distance.
type Distance struct {
	Inserts    int
	Deletes    int
	Replaces   int
	Transposes int
}

// NewDistance ports Distance(int inserts, int deletes, int replaces, int transposes).
func NewDistance(inserts, deletes, replaces, transposes int) Distance {
	return Distance{Inserts: inserts, Deletes: deletes, Replaces: replaces, Transposes: transposes}
}

// Insert ports Distance.insert().
func (d Distance) Insert() Distance {
	return Distance{d.Inserts + 1, d.Deletes, d.Replaces, d.Transposes}
}

// Delete ports Distance.delete().
func (d Distance) Delete() Distance {
	return Distance{d.Inserts, d.Deletes + 1, d.Replaces, d.Transposes}
}

// Replace ports Distance.replace().
func (d Distance) Replace() Distance {
	return Distance{d.Inserts, d.Deletes, d.Replaces + 1, d.Transposes}
}

// Transpose ports Distance.transpose().
func (d Distance) Transpose() Distance {
	return Distance{d.Inserts, d.Deletes, d.Replaces, d.Transposes + 1}
}

// Track ports Distance.track(EditOperation).
func (d Distance) Track(operation EditOperation) Distance {
	switch operation.(type) {
	case *InsertOp:
		return d.Insert()
	case *DeleteOp:
		return d.Delete()
	case *ReplaceOp:
		return d.Replace()
	case *TransposeOp:
		return d.Transpose()
	default:
		panic(fmt.Sprintf("Unknown operation: %v", operation))
	}
}

// Value ports Distance.value().
func (d Distance) Value() int {
	return d.Inserts + d.Deletes + d.Replaces + d.Transposes
}

// Equals ports Distance.equals (field equality).
func (d Distance) Equals(other Distance) bool {
	return d.Inserts == other.Inserts &&
		d.Deletes == other.Deletes &&
		d.Replaces == other.Replaces &&
		d.Transposes == other.Transposes
}

// String ports Distance.toString (ToStringBuilder field style without class@hash prefix).
func (d Distance) String() string {
	return fmt.Sprintf("[value=%d,inserts=%d,deletes=%d,replaces=%d,transposes=%d]",
		d.Value(), d.Inserts, d.Deletes, d.Replaces, d.Transposes)
}

// EditOperation ports DetailedDamerauLevenstheinDistance.EditOperation.
// Apply returns ok=false when Java would return null.
type EditOperation interface {
	Apply(s string) (string, bool)
}

type editOpBase struct {
	random *rand.Rand
}

func (o *editOpBase) rng() *rand.Rand {
	if o.random == nil {
		// Java: new Random() — non-deterministic default.
		o.random = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return o.random
}

// seedRandom ports EditOperation(long seed) for tests / deterministic apply.
func (o *editOpBase) seedRandom(seed int64) {
	o.random = rand.New(rand.NewSource(seed))
}

// randomChar ports DetailedDamerauLevenstheinDistance.randomChar.
func randomChar(r *rand.Rand) uint16 {
	return uint16(int('a') + r.Intn(26))
}

// DeleteOp ports DetailedDamerauLevenstheinDistance.Delete.
type DeleteOp struct{ editOpBase }

// NewDelete ports new Delete().
func NewDelete() *DeleteOp { return &DeleteOp{} }

// NewDeleteSeeded ports EditOperation(long seed) applied to Delete (subclass has no seed ctor in Java;
// used for deterministic twin tests).
func NewDeleteSeeded(seed int64) *DeleteOp {
	o := &DeleteOp{}
	o.seedRandom(seed)
	return o
}

// Apply ports Delete.apply.
func (o *DeleteOp) Apply(s string) (string, bool) {
	chars := utf16.Encode([]rune(s))
	if len(chars) <= 1 {
		return "", false
	}
	i := o.rng().Intn(len(chars))
	var out []uint16
	if i == 0 {
		out = chars[1:]
	} else if i == len(chars)-1 {
		out = chars[:i]
	} else {
		out = make([]uint16, 0, len(chars)-1)
		out = append(out, chars[:i]...)
		out = append(out, chars[i+1:]...)
	}
	return string(utf16.Decode(out)), true
}

// TransposeOp ports DetailedDamerauLevenstheinDistance.Transpose.
type TransposeOp struct{ editOpBase }

// NewTranspose ports new Transpose().
func NewTranspose() *TransposeOp { return &TransposeOp{} }

// NewTransposeSeeded is a deterministic helper (see NewDeleteSeeded).
func NewTransposeSeeded(seed int64) *TransposeOp {
	o := &TransposeOp{}
	o.seedRandom(seed)
	return o
}

// Apply ports Transpose.apply.
func (o *TransposeOp) Apply(s string) (string, bool) {
	chars := utf16.Encode([]rune(s))
	if len(chars) <= 1 {
		return "", false
	}
	i := o.rng().Intn(len(chars) - 1)
	// Java: String transposed = "" + s.charAt(i + 1) + s.charAt(i);
	transposed := []uint16{chars[i+1], chars[i]}
	var out []uint16
	if i == 0 {
		out = append(transposed, chars[2:]...)
	} else if i == len(chars)-1 {
		// Dead branch in Java: i is always in [0, length-2]; kept for bug-for-bug structure.
		out = append(append([]uint16{}, chars[:i]...), transposed...)
	} else {
		out = make([]uint16, 0, len(chars))
		out = append(out, chars[:i]...)
		out = append(out, transposed...)
		out = append(out, chars[i+2:]...)
	}
	return string(utf16.Decode(out)), true
}

// InsertOp ports DetailedDamerauLevenstheinDistance.Insert.
type InsertOp struct{ editOpBase }

// NewInsert ports new Insert().
func NewInsert() *InsertOp { return &InsertOp{} }

// NewInsertSeeded is a deterministic helper (see NewDeleteSeeded).
func NewInsertSeeded(seed int64) *InsertOp {
	o := &InsertOp{}
	o.seedRandom(seed)
	return o
}

// Apply ports Insert.apply.
func (o *InsertOp) Apply(s string) (string, bool) {
	chars := utf16.Encode([]rune(s))
	i := o.rng().Intn(len(chars) + 1)
	c := randomChar(o.rng())
	var out []uint16
	if i == 0 {
		out = append([]uint16{c}, chars...)
	} else if i == len(chars) {
		out = append(append([]uint16{}, chars...), c)
	} else {
		out = make([]uint16, 0, len(chars)+1)
		out = append(out, chars[:i]...)
		out = append(out, c)
		out = append(out, chars[i:]...)
	}
	return string(utf16.Decode(out)), true
}

// ReplaceOp ports DetailedDamerauLevenstheinDistance.Replace.
type ReplaceOp struct{ editOpBase }

// NewReplace ports new Replace().
func NewReplace() *ReplaceOp { return &ReplaceOp{} }

// NewReplaceSeeded is a deterministic helper (see NewDeleteSeeded).
func NewReplaceSeeded(seed int64) *ReplaceOp {
	o := &ReplaceOp{}
	o.seedRandom(seed)
	return o
}

// Apply ports Replace.apply.
func (o *ReplaceOp) Apply(s string) (string, bool) {
	chars := utf16.Encode([]rune(s))
	if len(chars) == 0 {
		return "", false
	}
	i := o.rng().Intn(len(chars))
	c := randomChar(o.rng())
	var out []uint16
	if i == 0 {
		out = append([]uint16{c}, chars[1:]...)
	} else if i == len(chars)-1 {
		out = append(append([]uint16{}, chars[:i]...), c)
	} else {
		out = make([]uint16, 0, len(chars))
		out = append(out, chars[:i]...)
		out = append(out, c)
		out = append(out, chars[i+1:]...)
	}
	return string(utf16.Decode(out)), true
}

// EditOperations ports DetailedDamerauLevenstheinDistance.editOperations
// (Insert, Replace, Transpose, Delete).
var EditOperations = []EditOperation{
	NewInsert(), NewReplace(), NewTranspose(), NewDelete(),
}

// editRandom ports the private static Random used by randomEdit().
var editRandom = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandomEdit ports DetailedDamerauLevenstheinDistance.randomEdit().
func RandomEdit() EditOperation {
	return EditOperations[editRandom.Intn(len(EditOperations))]
}

// Compare ports DetailedDamerauLevenstheinDistance.compare.
// Java null arguments throw NPE; Go has no null strings.
func Compare(s1, s2 string) Distance {
	if s1 == s2 {
		return Distance{}
	}

	// Java String length / charAt = UTF-16 code units.
	a := utf16.Encode([]rune(s1))
	b := utf16.Encode([]rune(s2))

	// INFinite distance is the max possible distance
	inf := len(a) + len(b)

	// Create and initialize the character array indices
	da := make(map[uint16]int)
	for _, ch := range a {
		da[ch] = 0
	}
	for _, ch := range b {
		da[ch] = 0
	}

	// Create the distance matrix H[0 .. s1.length+1][0 .. s2.length+1]
	h := make([][]Distance, len(a)+2)
	for i := range h {
		h[i] = make([]Distance, len(b)+2)
		// fill with empty values (new Distance())
		for j := range h[i] {
			h[i][j] = Distance{}
		}
	}

	// initialize the left and top edges of H
	// TODO understand initialization (kept from Java)
	for i := 0; i <= len(a); i++ {
		h[i+1][0] = NewDistance(inf, 0, 0, 0)
		h[i+1][1] = NewDistance(i, 0, 0, 0)
	}
	for j := 0; j <= len(b); j++ {
		h[0][j+1] = NewDistance(inf, 0, 0, 0)
		h[1][j+1] = NewDistance(j, 0, 0, 0)
	}

	// fill in the distance matrix H
	// look at each character in s1
	for i := 1; i <= len(a); i++ {
		db := 0
		// look at each character in b
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

			// Java: IntStream.of(...).min()
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

			// TODO: think about order, multiple cases might be true, which operations should be preferred?
			// Java if/else chain prefers: substitution > insertion > deletion > transpose
			if minV == substitution {
				if cost == 1 {
					h[i+1][j+1] = h[i][j].Replace()
				} else {
					h[i+1][j+1] = h[i][j]
				}
			} else if minV == insertion {
				h[i+1][j+1] = h[i+1][j].Insert()
			} else if minV == deletion {
				h[i+1][j+1] = h[i][j+1].Delete()
			} else if minV == transpose {
				transposeCost := (i - i1 - 1) + 1 + (j - j1 - 1)
				value := h[i1][j1]
				for k := 0; k < transposeCost; k++ {
					value = value.Transpose()
				}
				h[i+1][j+1] = value
			}
		}
		da[a[i-1]] = i
	}

	return h[len(a)+1][len(b)+1]
}
