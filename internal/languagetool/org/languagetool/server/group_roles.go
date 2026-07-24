package server

import "strings"

// GroupRole ports org.languagetool.server.GroupRoles.
type GroupRole string

const (
	GroupRoleExistingMember GroupRole = "EXISTING_MEMBER"
	GroupRoleMember         GroupRole = "MEMBER"
	GroupRoleOwner          GroupRole = "OWNER"
	GroupRoleAdmin          GroupRole = "ADMIN"
	GroupRoleEditor         GroupRole = "EDITOR"

	GroupRoleSeparator = ","
)

func EncodeGroupRoles(roles []GroupRole) string {
	parts := make([]string, len(roles))
	for i, r := range roles {
		parts[i] = string(r)
	}
	return strings.Join(parts, GroupRoleSeparator)
}

func DecodeGroupRoles(value string) []GroupRole {
	// Java GroupRoles.decode: value.split(SEPARATOR) then valueOf — no trim.
	// Empty value: split(",") on "" yields {""} in Java → valueOf("") throws;
	// Go fail-closed: nil for empty storage string.
	if value == "" {
		return nil
	}
	parts := strings.Split(value, GroupRoleSeparator)
	out := make([]GroupRole, 0, len(parts))
	for _, p := range parts {
		// Keep empty segments? Java valueOf("") throws IllegalArgumentException.
		// Fail-closed skip empty rather than invent a role.
		if p == "" {
			continue
		}
		out = append(out, GroupRole(p))
	}
	return out
}

// HasGroupPermissions reports whether role string includes any of the required roles.
func HasGroupPermissions(role string, required ...GroupRole) bool {
	if role == "" || len(required) == 0 {
		return false
	}
	have := map[GroupRole]struct{}{}
	for _, r := range DecodeGroupRoles(role) {
		have[r] = struct{}{}
	}
	for _, r := range required {
		if _, ok := have[r]; ok {
			return true
		}
	}
	return false
}
