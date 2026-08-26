package patterns

// Twin of StartupTimePerformanceTest — pattern builder startup smoke.
import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Port of StartupTimePerformanceTest (no @Test)
func TestStartupTimePerformance_NoTests(t *testing.T) {
	start := time.Now()
	_ = NewPatternTokenBuilder().Token("x").Negate().Min(0).Max(2).Build()
	require.Less(t, time.Since(start), time.Second)
}
