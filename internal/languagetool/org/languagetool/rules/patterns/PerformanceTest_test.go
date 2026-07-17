package patterns

// Twin of PerformanceTest (Java no @Test / interactive) — lightweight pattern smoke.
import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Port of PerformanceTest (no @Test)
func TestPerformance_NoTests(t *testing.T) {
	// soft: building simple PatternTokens stays under a trivial budget
	start := time.Now()
	for i := 0; i < 1000; i++ {
		pt := NewPatternTokenBuilder().Token("test").Build()
		require.NotNil(t, pt)
	}
	require.Less(t, time.Since(start), 2*time.Second)
}
