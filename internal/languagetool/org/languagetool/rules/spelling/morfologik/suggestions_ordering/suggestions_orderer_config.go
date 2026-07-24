package suggestions_ordering

import (
	"strconv"
	"strings"
	"sync"
)

// PROP_NAME ports SuggestionsOrdererConfig.PROP_NAME.
// Java: private static final String PROP_NAME = "enableMLSuggestionsOrdering";
const propName = "enableMLSuggestionsOrdering"

var (
	configMu   sync.RWMutex
	ngramsPath string
	// systemProperties twins JVM System properties used by this type
	// (System.getProperty / System.setProperty), process-global in-map.
	systemProperties map[string]string
)

// SetNgramsPath ports SuggestionsOrdererConfig.setNgramsPath.
func SetNgramsPath(path string) {
	configMu.Lock()
	defer configMu.Unlock()
	ngramsPath = path
}

// GetNgramsPath ports SuggestionsOrdererConfig.getNgramsPath.
func GetNgramsPath() string {
	configMu.RLock()
	defer configMu.RUnlock()
	return ngramsPath
}

// IsMLSuggestionsOrderingEnabled ports SuggestionsOrdererConfig.isMLSuggestionsOrderingEnabled.
// Java: Boolean.parseBoolean(System.getProperty(PROP_NAME, "false"))
func IsMLSuggestionsOrderingEnabled() bool {
	return parseBoolean(getSystemProperty(propName, "false"))
}

// SetMLSuggestionsOrderingEnabled ports SuggestionsOrdererConfig.setMLSuggestionsOrderingEnabled.
// Java: System.setProperty(PROP_NAME, String.valueOf(enabled))
func SetMLSuggestionsOrderingEnabled(enabled bool) {
	setSystemProperty(propName, strconv.FormatBool(enabled))
}

// getSystemProperty ports System.getProperty(key, defaultValue).
func getSystemProperty(key, defaultValue string) string {
	configMu.RLock()
	defer configMu.RUnlock()
	if systemProperties == nil {
		return defaultValue
	}
	v, ok := systemProperties[key]
	if !ok {
		return defaultValue
	}
	return v
}

// setSystemProperty ports System.setProperty(key, value).
func setSystemProperty(key, value string) {
	configMu.Lock()
	defer configMu.Unlock()
	if systemProperties == nil {
		systemProperties = make(map[string]string)
	}
	systemProperties[key] = value
}

// clearSystemProperty ports System.clearProperty(key) for tests / reset.
func clearSystemProperty(key string) {
	configMu.Lock()
	defer configMu.Unlock()
	if systemProperties == nil {
		return
	}
	delete(systemProperties, key)
}

// parseBoolean ports java.lang.Boolean.parseBoolean:
// true iff s is non-null and equals "true" ignoring case; otherwise false.
// (s is never null in Go; empty string and garbage → false.)
func parseBoolean(s string) bool {
	return strings.EqualFold(s, "true")
}
