package remote

// Twin of RemoteServerTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteServer_ToStringOutput(t *testing.T) {
	s := NewRemoteServerFull("Languagetool", "4.5-SNAPSHOT", "2019-02-05 17:54")
	require.Equal(t, "Languagetool/4.5-SNAPSHOT/2019-02-05 17:54", s.String())
}
