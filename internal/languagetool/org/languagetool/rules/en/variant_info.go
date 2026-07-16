package en

// VariantInfo ports org.languagetool.rules.en.VariantInfo.
type VariantInfo struct {
	VariantName  string
	OtherVariant string
}

func NewVariantInfo(variantName, otherVariant string) VariantInfo {
	if variantName == "" || otherVariant == "" {
		panic("variant fields required")
	}
	return VariantInfo{VariantName: variantName, OtherVariant: otherVariant}
}

func (v VariantInfo) GetVariantName() string  { return v.VariantName }
func (v VariantInfo) GetOtherVariant() string { return v.OtherVariant }
