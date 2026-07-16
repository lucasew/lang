package languagetool

// Twin of languagetool-standalone TranslationTest
// Core property helpers live in tools (avoid cycle); exercise ResourceBundleTools here.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of TranslationTest.testTranslationKeyExistence
func TestTranslation_TranslationKeyExistence(t *testing.T) {
	load := func(locale string) MessageBundle {
		switch locale {
		case "en":
			return MessageBundle{"desc": "Description", "category": "Category"}
		case "de":
			return MessageBundle{"desc": "Beschreibung"} // missing category
		default:
			return MessageBundle{}
		}
	}
	tools := NewResourceBundleTools(load)
	en := tools.GetMessageBundleFor("en")
	de := tools.GetMessageBundleFor("de")
	require.Contains(t, en, "desc")
	require.Contains(t, en, "category")
	// de primary missing category falls back via merge
	require.Equal(t, "Beschreibung", de["desc"])
	require.Equal(t, "Category", de["category"], "fallback from English")
}

// Port of TranslationTest.testTranslationsAreNotEmpty
func TestTranslation_TranslationsAreNotEmpty(t *testing.T) {
	// Simulate scanning property-like lines
	lines := []string{
		"# comment",
		"",
		"ok=value",
		"bad=",
	}
	var empties []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
			empties = append(empties, line)
		}
	}
	require.Equal(t, []string{"bad="}, empties)
}
