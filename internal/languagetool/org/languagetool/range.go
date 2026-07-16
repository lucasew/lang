package languagetool

// Range ports org.languagetool.Range — text span with guessed language.
type Range struct {
	FromPos int
	ToPos   int
	Lang    string
}

func NewRange(fromPos, toPos int, lang string) Range {
	if lang == "" {
		panic("lang required")
	}
	return Range{FromPos: fromPos, ToPos: toPos, Lang: lang}
}

func (r Range) GetFromPos() int { return r.FromPos }
func (r Range) GetToPos() int   { return r.ToPos }
func (r Range) GetLang() string { return r.Lang }

func (r Range) Equal(o Range) bool {
	return r.FromPos == o.FromPos && r.ToPos == o.ToPos && r.Lang == o.Lang
}
