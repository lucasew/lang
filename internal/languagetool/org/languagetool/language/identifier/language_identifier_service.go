package identifier

import "sync"

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

// GetSimpleLanguageIdentifier returns existing simple or creates via factory once.
func (s *LanguageIdentifierService) GetSimpleLanguageIdentifier(factory func() LanguageIdentifier) LanguageIdentifier {
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

// CanLanguageBeDetected is true if langCode is in supported or additional lists.
func CanLanguageBeDetected(langCode string, supported, additional []string) bool {
	for _, s := range supported {
		if s == langCode {
			return true
		}
	}
	for _, s := range additional {
		if s == langCode {
			return true
		}
	}
	return false
}
