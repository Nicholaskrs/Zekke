package helpers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)

	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		1,
		64*1024,
		4,
		32,
	)

	return fmt.Sprintf(
		"argon2id$v=19$m=65536,t=1,p=4$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func VerifyPassword(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 {
		return false
	}

	// argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	if parts[0] != "argon2id" || parts[1] != "v=19" {
		return false
	}

	// Parse m=65536,t=1,p=4
	params := strings.Split(parts[2], ",")
	if len(params) != 3 {
		return false
	}

	memory, err := strconv.ParseUint(strings.TrimPrefix(params[0], "m="), 10, 32)
	if err != nil {
		return false
	}

	iterations, err := strconv.ParseUint(strings.TrimPrefix(params[1], "t="), 10, 32)
	if err != nil {
		return false
	}

	parallelism, err := strconv.ParseUint(strings.TrimPrefix(params[2], "p="), 10, 8)
	if err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		uint32(iterations),
		uint32(memory),
		uint8(parallelism),
		uint32(len(expectedHash)),
	)

	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1
}
