package x

import (
	"crypto/sha256"
	"encoding/base64"
)

// ComputeEtag computes strong ETag for the given data.
func ComputeEtag(data ...[]byte) string {
	hash := sha256.New()
	for _, d := range data {
		_, _ = hash.Write(d)
	}
	return `"` + base64.RawURLEncoding.EncodeToString(hash.Sum(nil)) + `"`
}
