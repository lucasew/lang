package server

// Twin of HTTPServerMultiLangLoadTest — multi-lang detect soft load.
import (
	"sync"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of HTTPServerMultiLangLoadTest (no @Test)
func TestHTTPServerMultiLangLoad_NoTests(t *testing.T) {
	samples := map[string]string{
		"en": "Hello world sample text.",
		"de": "Die Größe des Hauses ist beachtlich.",
		"uk": "Це українська мова з ї.",
		"fr": "Bonjour le monde et merci.",
		"es": "Hola mundo de prueba.",
	}
	var wg sync.WaitGroup
	errCh := make(chan string, len(samples)*4)
	for code, text := range samples {
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func(c, tx string) {
				defer wg.Done()
				lt := languagetool.NewJLanguageTool(c)
				if len(lt.Analyze(tx)) == 0 {
					errCh <- "empty analyze " + c
					return
				}
				// detect heuristic (may not match code exactly for all)
				_ = DetectLanguageOfString(tx, []string{c}, nil)
			}(code, text)
		}
	}
	wg.Wait()
	close(errCh)
	for e := range errCh {
		t.Error(e)
	}
	// preferred variant path
	require.Equal(t, "en-GB", DetectLanguageOfString("Hello", []string{"en-GB"}, func(string) string { return "en" }))
}
