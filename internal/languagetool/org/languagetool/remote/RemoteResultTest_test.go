package remote

// Twin of RemoteResultTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteResult_GetLanguageDetectedCodeOutput(t *testing.T) {
	server := NewRemoteServerFull("LanguageTool", "", "")
	res := NewRemoteResultDetected("English", "en", "en", "English", nil, nil, server)
	require.Equal(t, "en", res.GetLanguageDetectedCode())
	require.Equal(t, "English", res.GetLanguageDetectedName())
}
