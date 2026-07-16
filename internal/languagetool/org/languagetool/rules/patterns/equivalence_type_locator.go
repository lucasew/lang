package patterns

// EquivalenceTypeLocator ports org.languagetool.rules.patterns.EquivalenceTypeLocator.
type EquivalenceTypeLocator struct {
	Feature string
	Type    string
}

func NewEquivalenceTypeLocator(feature, typ string) EquivalenceTypeLocator {
	return EquivalenceTypeLocator{Feature: feature, Type: typ}
}

func (e EquivalenceTypeLocator) Equal(o EquivalenceTypeLocator) bool {
	return e.Feature == o.Feature && e.Type == o.Type
}
