package en

// AvsAnData is the Java-name twin for a/an determiner word lists.
type AvsAnData struct{}

func (AvsAnData) RequiresA(word string) bool {
	loadAvsAnData()
	return wordsRequireA[word]
}

func (AvsAnData) RequiresAn(word string) bool {
	loadAvsAnData()
	return wordsRequireAn[word]
}

// UnitConversionRuleImperial is the Java-name twin of the imperial unit conversion variant.
type UnitConversionRuleImperial struct {
	*UnitConversionRule
}

func NewUnitConversionRuleImperialTyped(messages map[string]string) *UnitConversionRuleImperial {
	return &UnitConversionRuleImperial{UnitConversionRule: NewUnitConversionRuleImperial(messages)}
}
