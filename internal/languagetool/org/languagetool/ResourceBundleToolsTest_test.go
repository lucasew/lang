package languagetool

// Twin of ResourceBundleToolsTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of ResourceBundleToolsTest.testGetMessageBundle
func TestResourceBundleTools_GetMessageBundle(t *testing.T) {
	load := func(locale string) MessageBundle {
		switch locale {
		case "en":
			return MessageBundle{"greeting": "Hello", "farewell": "Bye"}
		case "fr":
			return MessageBundle{"greeting": "Bonjour"}
		default:
			return MessageBundle{}
		}
	}
	tools := NewResourceBundleTools(load)
	en := tools.GetMessageBundle()
	require.NotNil(t, en)
	require.Equal(t, "Hello", en["greeting"])
	fr := tools.GetMessageBundleFor("fr")
	require.Equal(t, "Bonjour", fr["greeting"])
	require.Equal(t, "Bye", fr["farewell"]) // English fallback
}
