package language

// French is the default France French variant (romance_variants.go).
var French = FrenchFrance

func NewFrench() FrenchVariant         { return FrenchFrance }
func NewCanadianFrench() FrenchVariant { return CanadianFrench }
func NewBelgianFrench() FrenchVariant  { return BelgianFrench }
func NewSwissFrench() FrenchVariant    { return SwissFrench }
