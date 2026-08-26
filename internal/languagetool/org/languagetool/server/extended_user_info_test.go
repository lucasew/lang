package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExtendedUserInfoAndDBAccess(t *testing.T) {
	from := time.Now().Add(-time.Hour)
	u := NewUserInfoEntry(5, "u@x.com")
	u.PremiumFrom = &from
	u.AddonToken = "tok"
	ext := FromUserInfoEntry(u, "User")
	require.Equal(t, "User", ext.Name)
	require.Equal(t, "tok", ext.AddonToken)
	back := ext.ToUserInfoEntry()
	require.Equal(t, int64(5), back.ID)

	oss := NewDatabaseAccessOpenSource()
	oss.Init()
	require.True(t, oss.IsReady())
	require.Nil(t, oss.GetUserByEmail("x"))
	oss.LogCheck(nil, 10, "uk", 1)
	require.GreaterOrEqual(t, DBLogger().QueueSize(), 1)
}
