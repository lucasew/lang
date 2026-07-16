package language

// English is the base English language (default variant American).
var English = AmericanEnglish

// Named twin constructors for Language subclasses.
func NewAmericanEnglish() EnglishVariant     { return AmericanEnglish }
func NewBritishEnglish() EnglishVariant      { return BritishEnglish }
func NewCanadianEnglish() EnglishVariant     { return CanadianEnglish }
func NewAustralianEnglish() EnglishVariant   { return AustralianEnglish }
func NewNewZealandEnglish() EnglishVariant   { return NewZealandEnglish }
func NewSouthAfricanEnglish() EnglishVariant { return SouthAfricanEnglish }
