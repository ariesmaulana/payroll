package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAndValidateDjangoPassword(t *testing.T) {

	password := "SecurePassword123!"
	salt := "customSalt123"
	iterations := 390000

	hashedPassword, err := CreateDjangoPBKDF2Password(password, salt, iterations)
	assert.Nil(t, err)

	verify := VerifyDjangoPBKDF2Password(password, hashedPassword)
	assert.True(t, verify)

}

func TestFailValidateDjangoPassword(t *testing.T) {

	password := "secure lah!"
	salt := "customSalt123"
	iterations := 390000

	hashedPassword, err := CreateDjangoPBKDF2Password(password, salt, iterations)
	assert.Nil(t, err)

	verify := VerifyDjangoPBKDF2Password("not match lah", hashedPassword)
	assert.False(t, verify)

}
