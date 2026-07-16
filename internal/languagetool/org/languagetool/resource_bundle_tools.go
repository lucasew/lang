package languagetool

// BundleLoader loads a MessageBundle for a locale code (e.g. "en", "en-US").
type BundleLoader func(locale string) MessageBundle

// ResourceBundleTools ports org.languagetool.ResourceBundleTools.
type ResourceBundleTools struct {
	Load BundleLoader
	// FallbackLocale defaults to "en".
	FallbackLocale string
}

func NewResourceBundleTools(load BundleLoader) *ResourceBundleTools {
	return &ResourceBundleTools{Load: load, FallbackLocale: "en"}
}

// GetMessageBundle returns a bundle for English with fallback (system default stub).
func (t *ResourceBundleTools) GetMessageBundle() MessageBundle {
	return t.GetMessageBundleFor("en")
}

// GetMessageBundleFor returns messages for lang, falling back to English via
// ResourceBundleWithFallback when both exist.
func (t *ResourceBundleTools) GetMessageBundleFor(lang string) MessageBundle {
	if t == nil || t.Load == nil {
		return MessageBundle{}
	}
	fbLocale := t.FallbackLocale
	if fbLocale == "" {
		fbLocale = "en"
	}
	primary := t.Load(lang)
	if primary == nil {
		return t.Load(fbLocale)
	}
	// if primary is empty map, try bare language
	if len(primary) == 0 {
		if i := indexByteRB(lang, '-'); i >= 0 {
			primary = t.Load(lang[:i])
		}
	}
	fallback := t.Load(fbLocale)
	if fallback == nil || len(fallback) == 0 {
		return primary
	}
	// return a merged view via ResourceBundleWithFallback.GetString semantics
	// as a flat map: primary overrides fallback
	out := MessageBundle{}
	for k, v := range fallback {
		out[k] = v
	}
	for k, v := range primary {
		if v != "" {
			out[k] = v
		}
	}
	return out
}

// GetMessageBundleWithFallbackStruct returns the pair for callers that need the wrapper.
func (t *ResourceBundleTools) GetMessageBundleWithFallbackStruct(lang string) *ResourceBundleWithFallback {
	if t == nil || t.Load == nil {
		return NewResourceBundleWithFallback(MessageBundle{}, MessageBundle{})
	}
	fbLocale := t.FallbackLocale
	if fbLocale == "" {
		fbLocale = "en"
	}
	return NewResourceBundleWithFallback(t.Load(lang), t.Load(fbLocale))
}

func indexByteRB(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
