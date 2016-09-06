package uuid

import (
	"crypto/rand"
	"fmt"
)

// Generate creates a new, version 4 uuid.
func NewUUID() (string, error) {
	// UUID representation compliant with specification described in RFC 4122.
	var u [16]byte

	if _, err := rand.Read(u[:]); err != nil {
		return "", err
	}

	// SetVersion sets version bits.
	u[6] = (u[6] & 0x0f) | (4 << 4)
	// SetVariant sets variant bits as described in RFC 4122.
	u[8] = (u[8] & 0xbf) | 0x80

	const format = "%08x-%04x-%04x-%04x-%012x"
	return fmt.Sprintf(format, u[:4], u[4:6], u[6:8], u[8:10], u[10:]), nil
}
