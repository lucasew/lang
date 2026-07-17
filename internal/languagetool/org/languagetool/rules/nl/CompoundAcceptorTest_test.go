package nl

// Twin of CompoundAcceptorTest — inject lists (Java @Ignore full speller deferred).
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of CompoundAcceptorTest.testAcceptCompound
func TestCompoundAcceptor_AcceptCompound(t *testing.T) {
	c := NewCompoundAcceptor()
	require.NoError(t, c.LoadNoS(strings.NewReader("auto\nfiets\n")))
	c.KnownWords["auto"] = struct{}{}
	c.KnownWords["weg"] = struct{}{}
	c.KnownWords["fiets"] = struct{}{}
	c.KnownWords["pad"] = struct{}{}
	require.True(t, c.Accept("autoweg"))
	require.True(t, c.Accept("TV-show")) // short/spelled hyphen form soft
	require.False(t, c.Accept("xyz"))
	require.False(t, c.Accept(""))
}

// Port of CompoundAcceptorTest.testAcceptCompoundInternal
func TestCompoundAcceptor_AcceptCompoundInternal(t *testing.T) {
	c := NewCompoundAcceptor()
	c.KnownWords["sport"] = struct{}{}
	c.KnownWords["wagen"] = struct{}{}
	require.True(t, c.Accept("sportwagen"))
	// too long
	long := strings.Repeat("a", 40)
	require.False(t, c.Accept(long))
}
