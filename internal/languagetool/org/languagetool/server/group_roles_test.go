package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupRoles_EncodeDecode(t *testing.T) {
	s := EncodeGroupRoles([]GroupRole{GroupRoleOwner, GroupRoleAdmin})
	require.Equal(t, "OWNER,ADMIN", s)
	got := DecodeGroupRoles(s)
	require.Equal(t, []GroupRole{GroupRoleOwner, GroupRoleAdmin}, got)
	require.True(t, HasGroupPermissions(s, GroupRoleAdmin))
	require.True(t, HasGroupPermissions(s, GroupRoleEditor, GroupRoleOwner))
	require.False(t, HasGroupPermissions(s, GroupRoleEditor))
	require.False(t, HasGroupPermissions("", GroupRoleAdmin))
	require.False(t, HasGroupPermissions("MEMBER")) // no required
}
