package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateUsername(t *testing.T) {
	t.Parallel()

	type testRows struct {
		username string
		success  bool
	}

	scenarios := []testRows{
		{
			username: "valid_user",
			success:  true, // Valid: Letters, numbers, underscores, and proper length.
		},
		{
			username: "user123",
			success:  true, // Valid: Only letters and numbers within length limits.
		},
		{
			username: "_invalid",
			success:  false, // Invalid: Starts with an underscore.
		},
		{
			username: "valid",
			success:  true, // valid: Ends with an underscore.
		},
		{
			username: "us",
			success:  false, // Invalid: Too short (minimum 5 characters required).
		},
		{
			username: "this_username_is_way_too_long",
			success:  false, // Invalid: Exceeds the 20-character limit.
		},
		{
			username: "valid123",
			success:  true, // Valid: Contains only letters and numbers.
		},
		{
			username: "user name",
			success:  false, // Invalid: Contains a space.
		},
		{
			username: "!@#$%^&*_",
			success:  false, // Invalid: invalid special character
		},
	}

	for _, v := range scenarios {
		v := v // Capture range variable to avoid parallel test issues
		t.Run(v.username, func(t *testing.T) {
			result := ValidateUsername(v.username)
			assert.Equal(t, v.success, result)
		})
	}
}
