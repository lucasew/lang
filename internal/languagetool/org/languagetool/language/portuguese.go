package language

// Portuguese defaults to Portugal Portuguese.
var Portuguese = PortugalPortuguese

func NewPortugalPortuguese() PortugueseVariant   { return PortugalPortuguese }
func NewBrazilianPortuguese() PortugueseVariant  { return BrazilianPortuguese }
func NewAngolaPortuguese() PortugueseVariant     { return AngolaPortuguese }
func NewMozambiquePortuguese() PortugueseVariant { return MozambiquePortuguese }
