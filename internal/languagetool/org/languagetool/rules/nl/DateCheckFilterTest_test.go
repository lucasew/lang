package nl

// Twin of DateCheckFilterTest (Dutch) — Java had no @Test; green helper smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of DateCheckFilterTest (surface)
func TestDateCheckFilter_NoTests(t *testing.T) {
	f := NewDateCheckFilter()
	d, err := f.GetDayOfWeekJava("maandag")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	m, err := f.GetMonth("mei")
	require.NoError(t, err)
	require.Equal(t, 5, m)
	require.Equal(t, "vrijdag", f.GetDayOfWeekName(2014, 8, 29))
}
