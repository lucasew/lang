package server

// Twin of RequestLimiterTest (Java king).
import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRequestLimiter_IsAccessOkay(t *testing.T) {
	l := NewRequestLimiter(2, 60)
	require.True(t, l.Allow("1.1.1.1"))
	require.True(t, l.Allow("1.1.1.1"))
	require.False(t, l.Allow("1.1.1.1"))
	require.True(t, l.Allow("2.2.2.2"))
	require.Equal(t, 2, l.Count("1.1.1.1"))
}

func TestRequestLimiter_IsAccessOkayWithFingerprintDisabled(t *testing.T) {
	l := NewRequestLimiter(1, 60)
	require.True(t, l.Allow("10.0.0.1"))
	require.False(t, l.Allow("10.0.0.1"))
}

func TestRequestLimiter_NilAllows(t *testing.T) {
	var l *RequestLimiter
	require.True(t, l.Allow("any"))
}

func assertOkay(t *testing.T, l *RequestLimiter, ip string, params map[string]string, header map[string][]string) {
	t.Helper()
	require.NoError(t, l.CheckAccess(ip, params, header, false))
}

func assertException(t *testing.T, l *RequestLimiter, ip string, params map[string]string, header map[string][]string) {
	t.Helper()
	err := l.CheckAccess(ip, params, header, false)
	require.Error(t, err)
	_, ok := err.(*TooManyRequestsError)
	require.True(t, ok, "want TooManyRequestsError, got %T %v", err, err)
}

// Twin of RequestLimiterTest.testIsAccessOkayWithByteLimitNoFingerprint
func TestRequestLimiter_IsAccessOkayWithByteLimitNoFingerprint(t *testing.T) {
	limiter := NewRequestLimiterFull(10, 35, 1, 0)
	firstIP := "192.168.10.1"
	secondIP := "192.168.10.2"
	firstHeader := map[string][]string{}
	secondHeader := map[string][]string{"User-Agent": {"Test"}}
	params := map[string]string{"text": "0123456789"}
	assertOkay(t, limiter, firstIP, params, firstHeader)      // 10
	assertOkay(t, limiter, firstIP, params, firstHeader)      // 20
	assertOkay(t, limiter, firstIP, params, firstHeader)      // 30
	assertException(t, limiter, firstIP, params, firstHeader) // 40
	assertException(t, limiter, firstIP, params, secondHeader)
	assertOkay(t, limiter, secondIP, params, firstHeader)
	assertOkay(t, limiter, secondIP, params, secondHeader)
	time.Sleep(1050 * time.Millisecond)
	assertOkay(t, limiter, firstIP, params, firstHeader)
	assertOkay(t, limiter, firstIP, params, secondHeader)
	assertOkay(t, limiter, secondIP, params, firstHeader)
	assertOkay(t, limiter, secondIP, params, secondHeader)
}

// Twin of RequestLimiterTest.testIsAccessOkayWithByteLimit
func TestRequestLimiter_IsAccessOkayWithByteLimit(t *testing.T) {
	limiter := NewRequestLimiterFull(10, 35, 1, 2)
	firstIP := "192.168.10.1"
	secondIP := "192.168.10.2"
	firstHeader := map[string][]string{}
	secondHeader := map[string][]string{"User-Agent": {"Test"}}
	params := map[string]string{"text": "0123456789"}
	assertOkay(t, limiter, firstIP, params, firstHeader)      // 10
	assertOkay(t, limiter, firstIP, params, firstHeader)      // 20
	assertOkay(t, limiter, firstIP, params, firstHeader)      // 30
	assertException(t, limiter, firstIP, params, firstHeader) // 40 fp1
	assertOkay(t, limiter, firstIP, params, secondHeader)     // new fingerprint
	assertOkay(t, limiter, firstIP, params, secondHeader)
	assertOkay(t, limiter, firstIP, params, secondHeader)
	assertException(t, limiter, firstIP, params, secondHeader) // 80 total IP, 40 fp2; ip limit 70
	assertOkay(t, limiter, secondIP, params, firstHeader)
	assertOkay(t, limiter, secondIP, params, secondHeader)
	time.Sleep(1050 * time.Millisecond)
	assertOkay(t, limiter, firstIP, params, firstHeader)
	assertOkay(t, limiter, firstIP, params, secondHeader)
	assertOkay(t, limiter, secondIP, params, firstHeader)
	assertOkay(t, limiter, secondIP, params, secondHeader)
}

// Twin of RequestLimiterTest.testTextLevelChecksCountLess
func TestRequestLimiter_TextLevelChecksCountLess(t *testing.T) {
	limiter := NewRequestLimiterFull(100, 35, 100, 2)
	firstIP := "192.168.10.1"
	firstHeader := map[string][]string{}
	params := map[string]string{"text": "0123456789"}
	assertOkay(t, limiter, firstIP, params, firstHeader) // 10
	assertOkay(t, limiter, firstIP, params, firstHeader) // 20
	assertOkay(t, limiter, firstIP, params, firstHeader) // 30
	params["mode"] = "textLevelOnly"
	assertOkay(t, limiter, firstIP, params, firstHeader) // +1 effective byte
	params["mode"] = "all"
	assertException(t, limiter, firstIP, params, firstHeader) // 41
	require.NoError(t, limiter.CheckAccess(firstIP, params, firstHeader, true))
}
