package language

// German is the default Germany German variant.
var German = GermanyGerman

func NewGermanyGerman() GermanVariant  { return GermanyGerman }
func NewAustrianGerman() GermanVariant { return AustrianGerman }
func NewSwissGerman() GermanVariant    { return SwissGerman }
