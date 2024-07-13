package generators

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

// logic to generate client ID for tw config
func GenerateUUID(email, id string) string {
	namespace := uuid.NewMD5(uuid.NameSpaceOID, []byte(email+id))
	return namespace.String()
}

// logic to generate encryption secret for tw config
func GenerateEncryptionSecret(uuidStr, email, id string) string {
	hash := sha256.New()
	hash.Write([]byte(uuidStr + email + id))
	return hex.EncodeToString(hash.Sum(nil))
}
