package server

import "time"

// ExtendedUserInfo ports org.languagetool.server.ExtendedUserInfo (API /users/me shape).
type ExtendedUserInfo struct {
	ID          int64      `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name,omitempty"`
	AddonToken  string     `json:"addon_token,omitempty"`
	APIKey      string     `json:"api_key,omitempty"`
	PremiumFrom *time.Time `json:"premium_from,omitempty"`
	PremiumTo   *time.Time `json:"premium_to,omitempty"`
	GroupID     *int64     `json:"group_id,omitempty"`
	GroupRole   string     `json:"group_role,omitempty"`
}

// ToUserInfoEntry maps extended JSON fields to UserInfoEntry.
func (e ExtendedUserInfo) ToUserInfoEntry() *UserInfoEntry {
	u := NewUserInfoEntry(e.ID, e.Email)
	u.AddonToken = e.AddonToken
	u.APIKey = e.APIKey
	u.PremiumFrom = e.PremiumFrom
	u.PremiumTo = e.PremiumTo
	u.UserGroup = e.GroupID
	u.GroupRole = e.GroupRole
	return u
}

// FromUserInfoEntry builds ExtendedUserInfo from UserInfoEntry.
func FromUserInfoEntry(u *UserInfoEntry, name string) ExtendedUserInfo {
	if u == nil {
		return ExtendedUserInfo{}
	}
	return ExtendedUserInfo{
		ID:          u.ID,
		Email:       u.Email,
		Name:        name,
		AddonToken:  u.AddonToken,
		APIKey:      u.APIKey,
		PremiumFrom: u.PremiumFrom,
		PremiumTo:   u.PremiumTo,
		GroupID:     u.UserGroup,
		GroupRole:   u.GroupRole,
	}
}
