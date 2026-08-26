package rules

// FileLineExpander ports the Java org.languagetool.rules.LineExpander interface.
// The CompoundRuleData pipeline uses the LineExpander func type; this interface
// is for implementations that need method-style expansion (e.g. Swiss ß→ss).
type FileLineExpander interface {
	ExpandLine(line string) []string
}

// FileLineExpanderFunc adapts a function to FileLineExpander.
type FileLineExpanderFunc func(line string) []string

func (f FileLineExpanderFunc) ExpandLine(line string) []string { return f(line) }

// AsLineExpander converts a FileLineExpander to the CompoundRuleData func type.
func AsLineExpander(e FileLineExpander) LineExpander {
	if e == nil {
		return nil
	}
	return e.ExpandLine
}
