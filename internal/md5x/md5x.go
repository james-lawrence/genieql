package md5x

import (
	"crypto/md5"
	"encoding/hex"
)

// Digest to md5 hex encoded string
func Digest(b []byte) string {
	d := md5.Sum(b)
	return hex.EncodeToString(d[:])
}

// String to md5 hex encoded string
func String(s string) string {
	return Digest([]byte(s))
}

// DigestX digest byte slice
func DigestX(b []byte) []byte {
	d := md5.Sum(b)
	return d[:]
}
