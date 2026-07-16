package languagetool

// MessageBundle is a minimal string-key message lookup (Java ResourceBundle surface).
type MessageBundle map[string]string

func (b MessageBundle) GetString(key string) string {
	if b == nil {
		return ""
	}
	return b[key]
}

// ResourceBundleWithFallback ports org.languagetool.ResourceBundleWithFallback.
type ResourceBundleWithFallback struct {
	Bundle   MessageBundle
	Fallback MessageBundle
}

func NewResourceBundleWithFallback(bundle, fallback MessageBundle) *ResourceBundleWithFallback {
	return &ResourceBundleWithFallback{Bundle: bundle, Fallback: fallback}
}

// GetString returns bundle value, or fallback if empty/missing.
func (r *ResourceBundleWithFallback) GetString(key string) string {
	s := r.Bundle.GetString(key)
	if stringsTrimSpaceEmpty(s) {
		return r.Fallback.GetString(key)
	}
	return s
}

func stringsTrimSpaceEmpty(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}
