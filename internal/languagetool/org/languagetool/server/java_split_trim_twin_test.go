package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin: HTTPServerConfig disabledRuleIds use split(",\\s*") not Fields/TrimSpace.
func TestSplitCommaOptionalASCIIWS(t *testing.T) {
	require.Equal(t, []string{"A", "B", "C"}, splitCommaOptionalASCIIWS("A, B,C"))
	require.Equal(t, []string{"A", "", "B"}, splitCommaOptionalASCIIWS("A,,B"))
	// NBSP after comma is not \\s — stays glued to next token
	require.Equal(t, []string{"A", "\u00a0B"}, splitCommaOptionalASCIIWS("A,\u00a0B"))
}

// Twin: TextChecker.getRuleValues — split on "," / ":" without trim.
func TestParseRuleValues_NoTrim(t *testing.T) {
	m := ParseRuleValues([]string{"RULE:1, OTHER:2"})
	// keys uppercased for Go map; spaces in id preserved from Java pair[0]
	require.Equal(t, "1", m["RULE"])
	require.Equal(t, "2", m[" OTHER"])
}

// Twin: GroupRoles.decode — no trim around commas.
func TestDecodeGroupRoles_NoTrim(t *testing.T) {
	// "ADMIN" ok; " ADMIN" would be invalid role name in Java valueOf — we keep surface
	roles := DecodeGroupRoles("ADMIN,EDITOR")
	require.Equal(t, []GroupRole{GroupRoleAdmin, GroupRoleEditor}, roles)
	roles = DecodeGroupRoles("ADMIN, EDITOR")
	require.Equal(t, []GroupRole{GroupRoleAdmin, GroupRole(" EDITOR")}, roles)
}
