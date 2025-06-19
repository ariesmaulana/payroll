package common

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// createDjangoPBKDF2Password generates a Django-compatible PBKDF2 password hash.
func CreateDjangoPBKDF2Password(password string, salt string, iterations int) (string, error) {
	if salt == "" {
		// Generate a random 16-byte salt and encode it in base64
		randomBytes := make([]byte, 16)
		_, err := rand.Read(randomBytes)
		if err != nil {
			return "", err
		}
		salt = base64.RawStdEncoding.EncodeToString(randomBytes) // Remove padding
	}

	// Perform the PBKDF2 hashing
	hashBytes := pbkdf2.Key([]byte(password), []byte(salt), iterations, 32, sha256.New)

	// Encode the hash in base64
	hashB64 := base64.StdEncoding.EncodeToString(hashBytes)

	// Construct the Django-compatible hash string
	hashString := fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", iterations, salt, hashB64)

	return hashString, nil
}

// verifyDjangoPBKDF2Password verifies a plaintext password against a Django-compatible PBKDF2 hash.
func VerifyDjangoPBKDF2Password(password, hashedPassword string) bool {
	// Split the Django-style hash
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 4 {
		return false
	}

	algorithm, iterStr, salt, hashB64 := parts[0], parts[1], parts[2], parts[3]
	if algorithm != "pbkdf2_sha256" {
		return false // Unsupported algorithm
	}

	// Convert iterations to int
	iterations, err := strconv.Atoi(iterStr)
	if err != nil {
		return false
	}

	// Recompute the hash using the given salt and iterations
	newHashBytes := pbkdf2.Key([]byte(password), []byte(salt), iterations, 32, sha256.New)
	newHashB64 := base64.StdEncoding.EncodeToString(newHashBytes)

	// Compare the stored hash with the newly computed hash securely
	return hmac.Equal([]byte(newHashB64), []byte(hashB64))
}

func GenerateRandomSalt() (string, error) {
	randomBytes := make([]byte, 16) // Django uses 16 bytes for salt
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the salt in base64 and remove padding "=" for Django compatibility
	salt := base64.RawStdEncoding.EncodeToString(randomBytes)
	return salt, nil
}
