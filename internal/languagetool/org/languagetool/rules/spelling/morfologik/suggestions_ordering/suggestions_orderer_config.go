package suggestions_ordering

import (
	"os"
	"strconv"
	"sync"
)

const propMLSuggestionsOrdering = "enableMLSuggestionsOrdering"

var (
	configMu   sync.RWMutex
	ngramsPath string
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
func IsMLSuggestionsOrderingEnabled() bool {
	v := os.Getenv(propMLSuggestionsOrdering)
	if v == "" {
		// also allow process env alternative used by tests
		v = os.Getenv("LT_" + propMLSuggestionsOrdering)
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false
	}
	return b
}

// SetMLSuggestionsOrderingEnabled ports SuggestionsOrdererConfig.setMLSuggestionsOrderingEnabled.
func SetMLSuggestionsOrderingEnabled(enabled bool) {
	_ = os.Setenv(propMLSuggestionsOrdering, strconv.FormatBool(enabled))
}
