package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// studioETag : empreinte courte du contenu servi (revalidation du cache).
func studioETag(data []byte) string {
	sum := sha256.Sum256(data)
	return `"` + hex.EncodeToString(sum[:8]) + `"`
}

// studioETagMatch : If-None-Match correspond-il à l'ETag courant ?
func studioETagMatch(header, etag string) bool {
	for _, p := range strings.Split(header, ",") {
		p = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(p), "W/"))
		if p == etag || p == "*" {
			return true
		}
	}
	return false
}
