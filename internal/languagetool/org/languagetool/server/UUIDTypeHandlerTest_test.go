package server

// Twin of UUIDTypeHandlerTest — full MyBatis/HSQL path deferred; green codec round-trip.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of UUIDTypeHandlerTest.testUUIDTypeHandler (binary codec surface)
func TestUUIDTypeHandler_UUIDTypeHandler(t *testing.T) {
	// nil / empty → zero bits
	z, err := BytesToUUIDBits(nil)
	require.NoError(t, err)
	require.Equal(t, uint64(0), z.MostSignificant)
	require.Equal(t, uint64(0), z.LeastSignificant)

	_, err = BytesToUUIDBits([]byte{1, 2, 3})
	require.Error(t, err)

	// known bits round-trip
	u := UUIDBits{MostSignificant: 0x0123456789abcdef, LeastSignificant: 0xfedcba9876543210}
	raw := UUIDBitsToBytes(u)
	require.Len(t, raw, 16)
	back, err := BytesToUUIDBits(raw)
	require.NoError(t, err)
	require.Equal(t, u, back)

	s := u.String()
	require.Contains(t, s, "-")
	parsed, err := ParseUUIDString(s)
	require.NoError(t, err)
	require.Equal(t, u.MostSignificant, parsed.MostSignificant)
	require.Equal(t, u.LeastSignificant, parsed.LeastSignificant)

	// standard random-looking UUID string
	std := "550e8400-e29b-41d4-a716-446655440000"
	p2, err := ParseUUIDString(std)
	require.NoError(t, err)
	require.NotEqual(t, uint64(0), p2.MostSignificant)
	require.Equal(t, std, p2.String())
}
