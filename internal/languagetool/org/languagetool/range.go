package languagetool

import "fmt"

// Range ports org.languagetool.Range — text span with guessed language.
type Range struct {
	FromPos int
	ToPos   int
	Lang    string
}

// NewRange ports Range(int, int, String). Java only rejects null lang
// (Objects.requireNonNull); empty string is allowed.
func NewRange(fromPos, toPos int, lang string) Range {
	return Range{FromPos: fromPos, ToPos: toPos, Lang: lang}
}

func (r Range) GetFromPos() int { return r.FromPos }
func (r Range) GetToPos() int   { return r.ToPos }
func (r Range) GetLang() string { return r.Lang }

func (r Range) Equal(o Range) bool {
	return r.FromPos == o.FromPos && r.ToPos == o.ToPos && r.Lang == o.Lang
}

// HashCode ports Range.hashCode (Objects.hash(fromPos, toPos, lang)).
func (r Range) HashCode() int {
	h := 1
	h = 31*h + r.FromPos
	h = 31*h + r.ToPos
	h = 31*h + stringHash(r.Lang)
	return h
}

func (r Range) String() string {
	return fmt.Sprintf("%d-%d:%s", r.FromPos, r.ToPos, r.Lang)
}
