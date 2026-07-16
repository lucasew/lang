package en

// UnitConversionRuleUS is the Java-name twin of the US unit conversion variant.
// Prefer NewUnitConversionRuleUS which returns *UnitConversionRule with Variant "us".
type UnitConversionRuleUS struct {
	*UnitConversionRule
}

// NewUnitConversionRuleUSTyped returns the named US wrapper (same behaviour as NewUnitConversionRuleUS).
func NewUnitConversionRuleUSTyped(messages map[string]string) *UnitConversionRuleUS {
	return &UnitConversionRuleUS{UnitConversionRule: NewUnitConversionRuleUS(messages)}
}
