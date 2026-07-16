package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceBundleTools(t *testing.T) {
	load := func(locale string) MessageBundle {
		switch locale {
		case "en":
			return MessageBundle{"hello": "Hello"}
		case "de":
			return MessageBundle{"hello": "Hallo"}
		default:
			return MessageBundle{}
		}
	}
	tools := NewResourceBundleTools(load)
	en := tools.GetMessageBundleFor("en")
	require.Equal(t, "Hello", en.GetString("hello"))
	de := tools.GetMessageBundleFor("de")
	require.Equal(t, "Hallo", de.GetString("hello"))
	// unknown lang falls back to en via merge
	xx := tools.GetMessageBundleFor("xx")
	require.Equal(t, "Hello", xx.GetString("hello"))
}
