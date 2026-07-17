package server

// Twin of HTTPServerLoadTest (Java interactive load) — concurrent detect/analyze soft.
import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of HTTPServerLoadTest (no @Test) — request-path smoke without live HTTP.
func TestHTTPServerLoad_NoTests(t *testing.T) {
	const workers = 16
	const perWorker = 10
	var ok atomic.Int32
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lt := languagetool.NewJLanguageTool("en")
			for i := 0; i < perWorker; i++ {
				sents := lt.Analyze("This is a load test sentence.")
				if len(sents) == 0 {
					return
				}
				code := DetectLanguageOfString("This is English text.", nil, nil)
				if code == "" {
					return
				}
				ok.Add(1)
			}
		}()
	}
	wg.Wait()
	require.Equal(t, int32(workers*perWorker), ok.Load())

	// limiter still usable under concurrent log
	lim := NewErrorRequestLimiter(100, 60)
	require.True(t, lim.WouldAccessBeOkay("10.0.0.1"))
}
