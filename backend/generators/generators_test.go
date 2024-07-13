package generators_test

import (
	"ccsync_backend/generators"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateUUID(t *testing.T) {
	email := "test@example.com"
	id := "12345"
	expectedUUID := uuid.NewMD5(uuid.NameSpaceOID, []byte(email+id)).String()

	uuidStr := generators.GenerateUUID(email, id)
	assert.Equal(t, expectedUUID, uuidStr)
}

func Test_GenerateEncryptionSecret(t *testing.T) {
	uuidStr := "uuid-test"
	email := "test@example.com"
	id := "12345"
	hash := sha256.New()
	hash.Write([]byte(uuidStr + email + id))
	expectedSecret := hex.EncodeToString(hash.Sum(nil))

	encryptionSecret := generators.GenerateEncryptionSecret(uuidStr, email, id)
	assert.Equal(t, expectedSecret, encryptionSecret)
}
