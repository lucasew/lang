package uk

// Twin of AbstractRuleTest — Java base class had no @Test; green agreement helpers smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of AbstractRuleTest (no tests in Java)
func TestAbstractRule_NoTests(t *testing.T) {
	require.NotEmpty(t, TokenAgreementAdjNounRuleID)
	require.NotEmpty(t, TokenAgreementPrepNounRuleID)
	require.NotEmpty(t, TokenAgreementVerbNounRuleID)
	require.NotNil(t, NewTokenAgreementAdjNounRule())
	require.NotNil(t, NewTokenAgreementPrepNounRule())
}
