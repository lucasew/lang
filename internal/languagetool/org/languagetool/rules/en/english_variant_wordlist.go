package en

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// english_variant_wordlist ports AbstractEnglishSpellerRule.loadWordlist for en/en-US-GB.txt.

const usGbVariantResource = "en/en-US-GB.txt"

var (
	usGbMu     sync.Mutex
	usGbMaps   [2]map[string]string
	usGbMapsOK [2]bool
)

// LoadUSGBVariantMap ports loadWordlist("en/en-US-GB.txt", column).
// column 0: key = US form (parts[0]), value = GB form (parts[1])
// column 1: key = GB form (parts[1]), value = US form (parts[0])
// Missing resource → empty map (fail-closed, no invent pairs).
func LoadUSGBVariantMap(column int) map[string]string {
	if column != 0 && column != 1 {
		return map[string]string{}
	}
	usGbMu.Lock()
	defer usGbMu.Unlock()
	if usGbMapsOK[column] {
		return usGbMaps[column]
	}
	m := map[string]string{}
	p := spelling.DiscoverSpellingResource(usGbVariantResource)
	if p == "" {
		usGbMaps[column] = m
		usGbMapsOK[column] = true
		return m
	}
	lines, err := spelling.LoadSpellingWordListFile(p)
	if err != nil {
		usGbMaps[column] = m
		usGbMapsOK[column] = true
		return m
	}
	for _, line := range lines {
		line = tools.JavaStringTrim(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) != 2 {
			// Java throws; skip malformed lines (fail-closed partial)
			continue
		}
		a, b := tools.JavaStringTrim(parts[0]), tools.JavaStringTrim(parts[1])
		if a == "" || b == "" {
			continue
		}
		// Java: words.put(parts[column].toLowerCase(), parts[column == 1 ? 0 : 1]);
		key := strings.ToLower(parts[column])
		val := tools.JavaStringTrim(parts[1-column])
		m[key] = val
	}
	usGbMaps[column] = m
	usGbMapsOK[column] = true
	return m
}
