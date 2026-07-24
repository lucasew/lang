package identifier

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// LanguageIdentifierService ports
// org.languagetool.language.identifier.LanguageIdentifierService as a process singleton.
type LanguageIdentifierService struct {
	mu                sync.Mutex
	defaultIdentifier LanguageIdentifier
	simpleIdentifier  LanguageIdentifier
}

// Instance is the package-level service (Java INSTANCE).
var Instance = &LanguageIdentifierService{}

// GetInitialized returns default, else simple, else nil.
func (s *LanguageIdentifierService) GetInitialized() LanguageIdentifier {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.defaultIdentifier != nil {
		return s.defaultIdentifier
	}
	return s.simpleIdentifier
}

// SetDefault sets the default identifier (for production wiring / tests).
func (s *LanguageIdentifierService) SetDefault(id LanguageIdentifier) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.defaultIdentifier = id
}

// SetSimple sets the simple identifier.
func (s *LanguageIdentifierService) SetSimple(id LanguageIdentifier) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.simpleIdentifier = id
}

// GetSimpleLanguageIdentifierWithFactory returns existing simple or creates via factory once.
func (s *LanguageIdentifierService) GetSimpleLanguageIdentifierWithFactory(factory func() LanguageIdentifier) LanguageIdentifier {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.simpleIdentifier == nil && factory != nil {
		s.simpleIdentifier = factory()
	}
	return s.simpleIdentifier
}

// Clear resets identifiers (Java clear for tests).
func (s *LanguageIdentifierService) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.defaultIdentifier = nil
	s.simpleIdentifier = nil
}

// ClearLanguageIdentifier ports clearLanguageIdentifier("default"|"simple"|"both").
func (s *LanguageIdentifierService) ClearLanguageIdentifier(typ string) *LanguageIdentifierService {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch typ {
	case "default":
		s.defaultIdentifier = nil
	case "simple":
		s.simpleIdentifier = nil
	case "both":
		s.defaultIdentifier = nil
		s.simpleIdentifier = nil
	}
	return s
}

// GetDefaultLanguageIdentifier ports getDefaultLanguageIdentifier(maxLength) without ngram/fasttext paths.
func (s *LanguageIdentifierService) GetDefaultLanguageIdentifier(maxLength int) LanguageIdentifier {
	return s.GetDefaultLanguageIdentifierFull(maxLength, "", "", "")
}

// GetDefaultLanguageIdentifierFull ports
// getDefaultLanguageIdentifier(maxLength, ngramLangIdentData, fasttextBinary, fasttextModel).
// Existing default is returned as-is (Java: only initialize once).
// Empty paths leave ngram/fasttext disabled.
func (s *LanguageIdentifierService) GetDefaultLanguageIdentifierFull(
	maxLength int,
	ngramLangIdentData, fasttextBinary, fasttextModel string,
) LanguageIdentifier {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.defaultIdentifier == nil {
		if maxLength <= 0 {
			maxLength = DefaultMaxLength
		}
		d := NewDefaultLanguageIdentifier(maxLength)
		// Java enableNgrams: RuntimeException on IOException
		if err := d.EnableNgramsFromPath(ngramLangIdentData); err != nil {
			panic("Could not load ngram data language identification from " + ngramLangIdentData + ": " + err.Error())
		}
		// Java enableFasttext: RuntimeException on IOException when both paths set
		if fasttextBinary != "" && fasttextModel != "" {
			if err := d.EnableFastTextFromPaths(fasttextBinary, fasttextModel); err != nil {
				panic("Could not start fasttext process for language identification @ " +
					fasttextBinary + " with model @ " + fasttextModel + ": " + err.Error())
			}
		}
		s.defaultIdentifier = d
	}
	return s.defaultIdentifier
}

// GetSimpleLanguageIdentifier ports getSimpleLanguageIdentifier(preferredLangCodes).
func (s *LanguageIdentifierService) GetSimpleLanguageIdentifier(preferred []string) LanguageIdentifier {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.simpleIdentifier == nil {
		s.simpleIdentifier = NewSimpleLanguageIdentifierWith(preferred, nil)
	}
	return s.simpleIdentifier
}

// CanLanguageBeDetected ports LanguageIdentifierService.canLanguageBeDetected.
// When supported is nil, uses Languages.isLanguageSupported via GlobalLanguages.
func CanLanguageBeDetected(langCode string, supported, additional []string) bool {
	// Java Languages static init before first call.
	languagetool.EnsureBuiltInLanguagesRegistered()
	if supported == nil {
		if languagetool.GlobalLanguages.IsLanguageSupported(langCode) {
			return true
		}
	} else {
		for _, s := range supported {
			if s == langCode {
				return true
			}
		}
	}
	for _, s := range additional {
		if s == langCode {
			return true
		}
	}
	return false
}

// CanLanguageBeDetected method ports INSTANCE.canLanguageBeDetected(lang, additional).
func (s *LanguageIdentifierService) CanLanguageBeDetected(langCode string, additional []string) bool {
	return CanLanguageBeDetected(langCode, nil, additional)
}
