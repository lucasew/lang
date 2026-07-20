package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetStream_MissingOpener(t *testing.T) {
	prev := AsStream
	AsStream = nil
	t.Cleanup(func() { AsStream = prev })
	_, err := GetStream("/foo")
	require.Error(t, err)
	require.Contains(t, err.Error(), "Could not load file from classpath: '/foo'")
}

func TestGetStream_AndHash(t *testing.T) {
	prev := AsStream
	body := "hello LT"
	AsStream = func(path string) (io.ReadCloser, error) {
		if path != "res/x.txt" {
			return nil, nil
		}
		return io.NopCloser(strings.NewReader(body)), nil
	}
	t.Cleanup(func() { AsStream = prev })

	rc, err := GetStream("res/x.txt")
	require.NoError(t, err)
	b, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, body, string(b))
	_ = rc.Close()

	sum := sha256.Sum256([]byte(body))
	hexSum := hex.EncodeToString(sum[:])
	rc2, err := GetStreamWithHash("res/x.txt", hexSum)
	require.NoError(t, err)
	b2, _ := io.ReadAll(rc2)
	require.Equal(t, body, string(b2))

	_, err = GetStreamWithHash("res/x.txt", "deadbeef")
	require.Error(t, err)
	require.Contains(t, err.Error(), "Checksum mismatch")
}

func TestIsExternSpeller(t *testing.T) {
	SetLinguisticServices(nil)
	require.False(t, IsExternSpeller())
	require.Nil(t, GetLinguisticServices())
	SetLinguisticServices(struct{}{})
	require.True(t, IsExternSpeller())
	require.NotNil(t, GetLinguisticServices())
	SetLinguisticServices(nil)
}
