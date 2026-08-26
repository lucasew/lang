package en

// VariantInfo ports org.languagetool.rules.en.VariantInfo.
type VariantInfo struct {
	VariantName  string
	OtherVariant string
}

// NewVariantInfo ports VariantInfo(String, String). Java uses Objects.requireNonNull
// (null only); empty strings are allowed.
func NewVariantInfo(variantName, otherVariant string) VariantInfo {
	return VariantInfo{VariantName: variantName, OtherVariant: otherVariant}
}

func (v VariantInfo) GetVariantName() string  { return v.VariantName }
func (v VariantInfo) OtherVariantName() string { return v.OtherVariant }

// GetOtherVariant is an alias for OtherVariantName (Java otherVariant()).
func (v VariantInfo) GetOtherVariant() string { return v.OtherVariant }
