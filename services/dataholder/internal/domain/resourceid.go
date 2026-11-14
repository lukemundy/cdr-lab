package domain

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// ResourceID has no key, because the internal ID is a random UUID that no data recipient gets.
func ResourceID(softwareID, customerNumber, internalID string) string {
	var data []byte

	for _, field := range []string{softwareID, customerNumber, internalID} {
		data = binary.BigEndian.AppendUint32(data, uint32(len(field)))
		data = append(data, field...)
	}

	sum := sha256.Sum256(data)
	b := sum[:16]
	// The version and variant bits of a version 8 UUID (RFC 9562, section 5.8).
	b[6] = b[6]&0x0f | 0x80
	b[8] = b[8]&0x3f | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
