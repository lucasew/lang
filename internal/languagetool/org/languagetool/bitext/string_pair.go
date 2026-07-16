package bitext

// StringPair ports org.languagetool.bitext.StringPair.
type StringPair struct {
	source string
	target string
}

func NewStringPair(source, target string) StringPair {
	return StringPair{source: source, target: target}
}

func (p StringPair) GetSource() string { return p.source }
func (p StringPair) GetTarget() string { return p.target }
