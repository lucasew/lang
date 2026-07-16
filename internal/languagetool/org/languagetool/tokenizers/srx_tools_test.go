package tokenizers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSrxTools(t *testing.T) {
	xml := `<?xml version="1.0"?>
	<srx>
	  <languagerule languagerulename="en"/>
	  <languagerule languagerulename="de"/>
	</srx>`
	doc := CreateSrxDocumentFromString(xml)
	require.Contains(t, doc.LanguageCodes, "en")
	require.Contains(t, doc.LanguageCodes, "de")

	doc2, err := CreateSrxDocumentFromReader(strings.NewReader(xml))
	require.NoError(t, err)
	require.Equal(t, doc.LanguageCodes, doc2.LanguageCodes)

	segs := TokenizeWithSrx("Hello world. Next.", doc, "en")
	require.NotEmpty(t, segs)
}
