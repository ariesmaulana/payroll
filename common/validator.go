package common

import "regexp"

// ValidateUsername checks if the username follows the rules:
// - 5 to 20 characters
// - Can contain letters, numbers, and underscores
// - Cannot start with an underscore
func ValidateUsername(username string) bool {
	if len(username) < 5 || len(username) > 20 {
		return false
	}

	// Updated regex: Disallow leading `_`, but allow it in the middle and at the end.
	re := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{1,19}$`)
	return re.MatchString(username)
}
