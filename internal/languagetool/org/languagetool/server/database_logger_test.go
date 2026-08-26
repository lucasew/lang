package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseLoggerAndGroups(t *testing.T) {
	l := NewDatabaseLogger(10)
	require.False(t, l.IsEnabled())
	l.Log(NewDatabasePingLogEntry(nil, nil))
	require.Equal(t, 0, l.QueueSize())
	l.Init()
	require.True(t, l.IsEnabled())
	uid := int64(1)
	l.Log(NewDatabasePingLogEntry(nil, &uid))
	l.Log(NewDatabaseCheckLogEntry(&uid, 5, "en", 0))
	require.Equal(t, 2, l.QueueSize())
	got := l.Poll(1)
	require.Len(t, got, 1)
	require.Equal(t, 1, l.QueueSize())

	DatabaseAccessInstance().InitOpenSource()
	require.True(t, DatabaseAccessInstance().IsReady())

	m := NewDBGroupMember(1, 2, 3, []GroupRole{GroupRoleMember, GroupRoleEditor})
	require.Contains(t, m.Role, "MEMBER")
	inv := NewDBInvite(9, 2, "a@b.c", "tok")
	require.Equal(t, "a@b.c", inv.Email)
}
