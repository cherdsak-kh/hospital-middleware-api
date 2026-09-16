package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordHashing(t *testing.T) {
	plainPassword := "SecretP@ss123"

	// Positive test: Hash password and verify correct match
	hash, err := HashPassword(plainPassword)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, CheckPassword(hash, plainPassword))

	// Negative test: Wrong password should return false
	assert.False(t, CheckPassword(hash, "WrongPassword"))
	assert.False(t, CheckPassword(hash, ""))

	// Negative test: Password exceeding 72 bytes triggers bcrypt error
	tooLongPassword := string(make([]byte, 100))
	tooLongHash, errTooLong := HashPassword(tooLongPassword)
	assert.Error(t, errTooLong)
	assert.Empty(t, tooLongHash)
}
