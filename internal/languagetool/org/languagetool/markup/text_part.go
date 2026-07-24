package markup

// TextPart ports org.languagetool.markup.TextPart.
type TextPartType int

const (
	TextPartText TextPartType = iota
	TextPartMarkup
	TextPartFakeContent
)

type TextPart struct {
	Part string
	Type TextPartType
}

func NewTextPart(part string, typ TextPartType) TextPart {
	return TextPart{Part: part, Type: typ}
}

func (p TextPart) String() string { return p.Part }
