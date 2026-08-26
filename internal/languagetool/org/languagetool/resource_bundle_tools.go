package languagetool

// BundleLoader loads a MessageBundle for a locale code (e.g. "en", "en-US").
// Java: JLanguageTool.getDataBroker().getResourceBundle(MESSAGE_BUNDLE, Locale).
type BundleLoader func(locale string) MessageBundle

// BundleLocale reports the language tag of a loaded bundle for isValidBundleFor.
// When LoadWithLocale is set, it is preferred over Load.
type BundleWithLocale struct {
	Bundle MessageBundle
	// Lang is bundle.getLocale().getLanguage() (e.g. "en").
	Lang string
}

// ResourceBundleTools ports org.languagetool.ResourceBundleTools.
// Static Java methods become methods on a tools instance with an injectable loader
// (data broker / MessageBundle not global in Go).
type ResourceBundleTools struct {
	Load BundleLoader
	// LoadWithLocale optionally returns locale language of the loaded bundle.
	LoadWithLocale func(locale string) BundleWithLocale
	// FallbackLocale defaults to "en" (Java Locale.ENGLISH).
	FallbackLocale string
	// SystemLocale for GetMessageBundle() (Java Locale.getDefault()); default "en".
	SystemLocale string
}

func NewResourceBundleTools(load BundleLoader) *ResourceBundleTools {
	return &ResourceBundleTools{Load: load, FallbackLocale: "en", SystemLocale: "en"}
}

func (t *ResourceBundleTools) load(locale string) BundleWithLocale {
	if t != nil && t.LoadWithLocale != nil {
		return t.LoadWithLocale(locale)
	}
	if t == nil || t.Load == nil {
		return BundleWithLocale{}
	}
	// Infer language from locale code when LoadWithLocale not set.
	lang := locale
	if i := indexByteRB(locale, '-'); i >= 0 {
		lang = locale[:i]
	}
	if i := indexByteRB(locale, '_'); i >= 0 {
		lang = locale[:i]
	}
	return BundleWithLocale{Bundle: t.Load(locale), Lang: lang}
}

// GetMessageBundle ports getMessageBundle() for the system default locale.
func (t *ResourceBundleTools) GetMessageBundle() MessageBundle {
	sys := "en"
	if t != nil && t.SystemLocale != "" {
		sys = t.SystemLocale
	}
	return t.getMessageBundleForLocale(sys)
}

// GetMessageBundleFor ports getMessageBundle(Language) using short code / variant code.
// langCode may be "de", "en-US", etc.
func (t *ResourceBundleTools) GetMessageBundleFor(langCode string) MessageBundle {
	return t.getMessageBundleForLocale(langCode)
}

func (t *ResourceBundleTools) getMessageBundleForLocale(langCode string) MessageBundle {
	if t == nil {
		return MessageBundle{}
	}
	fbLocale := t.FallbackLocale
	if fbLocale == "" {
		fbLocale = "en"
	}
	// try/catch MissingResourceException → English
	defer func() { _ = recover() }()

	// Java: bundle for localeWithCountryAndVariant, then validate language
	bundle := t.load(langCode)
	if !isValidBundleFor(langCode, bundle) {
		// try bare language locale
		bare := langCode
		if i := indexByteRB(langCode, '-'); i >= 0 {
			bare = langCode[:i]
		} else if i := indexByteRB(langCode, '_'); i >= 0 {
			bare = langCode[:i]
		}
		bundle = t.load(bare)
		if !isValidBundleFor(langCode, bundle) {
			// default variant path not fully available without Language twin;
			// leave bundle as-is (may be empty).
		}
	}
	if bundle.Bundle == nil {
		return t.load(fbLocale).Bundle
	}
	fallback := t.load(fbLocale)
	if fallback.Bundle == nil {
		return bundle.Bundle
	}
	return flattenWithFallback(bundle.Bundle, fallback.Bundle)
}

// isValidBundleFor ports private isValidBundleFor(Language, ResourceBundle):
// lang.getLocale().getLanguage().equals(bundle.getLocale().getLanguage()).
func isValidBundleFor(langCode string, bundle BundleWithLocale) bool {
	if bundle.Bundle == nil || len(bundle.Bundle) == 0 {
		return false
	}
	want := langCode
	if i := indexByteRB(langCode, '-'); i >= 0 {
		want = langCode[:i]
	} else if i := indexByteRB(langCode, '_'); i >= 0 {
		want = langCode[:i]
	}
	return want == bundle.Lang
}

func flattenWithFallback(primary, fallback MessageBundle) MessageBundle {
	out := MessageBundle{}
	for k, v := range fallback {
		out[k] = v
	}
	for k, v := range primary {
		if !stringsTrimSpaceEmpty(v) {
			out[k] = v
		}
	}
	return out
}

// GetMessageBundleWithFallbackStruct returns the pair for callers that need the wrapper.
func (t *ResourceBundleTools) GetMessageBundleWithFallbackStruct(lang string) *ResourceBundleWithFallback {
	if t == nil {
		return NewResourceBundleWithFallback(MessageBundle{}, MessageBundle{})
	}
	fbLocale := t.FallbackLocale
	if fbLocale == "" {
		fbLocale = "en"
	}
	return NewResourceBundleWithFallback(t.load(lang).Bundle, t.load(fbLocale).Bundle)
}

func indexByteRB(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
