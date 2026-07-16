package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/ErrorRequestLimiterTest.java
import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Port of ErrorRequestLimiterTest.testErrorLimiter
func TestErrorRequestLimiter_ErrorLimiter(t *testing.T) {
	limiter := NewErrorRequestLimiter(3, 1)
	ip1 := "192.168.0.1"
	ip2 := "192.168.0.2"
	require.True(t, limiter.WouldAccessBeOkay(ip1))
	require.True(t, limiter.WouldAccessBeOkay(ip2))
	limiter.LogAccess(ip1)
	limiter.LogAccess(ip1)
	limiter.LogAccess(ip1)
	limiter.LogAccess(ip1)
	require.False(t, limiter.WouldAccessBeOkay(ip1))
	require.True(t, limiter.WouldAccessBeOkay(ip2))
	time.Sleep(1100 * time.Millisecond)
	require.True(t, limiter.WouldAccessBeOkay(ip1))
}
