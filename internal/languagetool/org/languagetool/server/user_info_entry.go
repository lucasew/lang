package server

import "time"

// UserInfoEntry ports org.languagetool.server.UserInfoEntry.
type UserInfoEntry struct {
	ID                int64
	Email             string
	PasswordHash      []byte
	AddonToken        string
	APIKey            string
	UserDictCacheSize *int64
	RequestsPerDay    *int64
	LimitEnforcement  LimitEnforcementMode
	ManagedAccounts   *int64
	PremiumFrom       *time.Time
	PremiumTo         *time.Time
	UserGroup         *int64
	GroupID           string // UUID string
	GroupRole         string
	DefaultDictionary string
	OptIn3rdPartyAIGrammarChecker bool
	OptIn3rdPartyAIParaphraser    bool
}

func NewUserInfoEntry(id int64, email string) *UserInfoEntry {
	return &UserInfoEntry{
		ID:               id,
		Email:            email,
		LimitEnforcement: LimitEnforcementDisabled,
	}
}

// HasPremium reports whether premium is active now (or open-ended).
func (u *UserInfoEntry) HasPremium() bool {
	if u == nil {
		return false
	}
	now := time.Now()
	if u.PremiumTo != nil && now.After(*u.PremiumTo) {
		return false
	}
	if u.PremiumFrom != nil && now.Before(*u.PremiumFrom) {
		return false
	}
	// Premium if either bound is set (active window) or PremiumTo is in the future.
	return u.PremiumFrom != nil || u.PremiumTo != nil
}
