// Package envx provides utility functions for extracting information from environment variables
package envx

import (
	"encoding/base64"
	"encoding/hex"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Int retrieve a integer flag from the environment, checks each key in order
// first to parse successfully is returned.
func Int(fallback int, keys ...string) int {
	return envval(fallback, func(s string) (int, error) {
		decoded, err := strconv.ParseInt(s, 10, 64)
		return int(decoded), errors.Wrapf(err, "integer '%s' is invalid", s)
	}, keys...)
}

// Boolean retrieve a boolean flag from the environment, checks each key in order
// first to parse successfully is returned.
func Boolean(fallback bool, keys ...string) bool {
	return envval(fallback, func(s string) (bool, error) {
		decoded, err := strconv.ParseBool(s)
		return decoded, errors.Wrapf(err, "boolean '%s' is invalid", s)
	}, keys...)
}

// Float64 retrieve a float64 flag from the environment, checks each key in order
// first to parse successfully is returned.
func Float64(fallback float64, keys ...string) float64 {
	return envval(fallback, func(s string) (float64, error) {
		decoded, err := strconv.ParseFloat(s, 64)
		return decoded, errors.Wrapf(err, "float64 '%s' is invalid", s)
	}, keys...)
}

// String retrieve a string value from the environment, checks each key in order
// first string found is returned.
func String(fallback string, keys ...string) string {
	return envval(fallback, func(s string) (string, error) {
		// we'll never receive an empty string because envval skips empty strings.
		return s, nil
	}, keys...)
}

// String retrieve a string value from the environment, checks each key in order
// first string found is returned.
func Strings(fallback []string, keys ...string) []string {
	return envval(fallback, func(s string) ([]string, error) {
		// we'll never receive an empty string because envval skips empty strings.
		return strings.Split(s, ","), nil
	}, keys...)
}

// Duration retrieves a time.Duration from the environment, checks each key in order
// first successful parse to a duration is returned.
func Duration(fallback time.Duration, keys ...string) time.Duration {
	return envval(fallback, func(s string) (time.Duration, error) {
		decoded, err := time.ParseDuration(s)
		return decoded, errors.Wrapf(err, "time.Duration '%s' is invalid", s)
	}, keys...)
}

// BytesFile treats the value in the provided environment keys as a file path.
func BytesFile(fallback []byte, keys ...string) []byte {
	return envval(fallback, func(s string) ([]byte, error) {
		decoded, err := os.ReadFile(s)
		return decoded, errors.Wrapf(err, "file path '%s' was inaccessible", s)
	}, keys...)
}

// BytesHex read value as a hex encoded string.
func BytesHex(fallback []byte, keys ...string) []byte {
	return envval(fallback, func(s string) ([]byte, error) {
		decoded, err := hex.DecodeString(s)
		return decoded, errors.Wrapf(err, "invalid hex encoded data '%s'", s)
	}, keys...)
}

// BytesB64 read value as a base64 encoded string
func BytesB64(fallback []byte, keys ...string) []byte {
	enc := base64.RawStdEncoding.WithPadding('=')
	return envval(fallback, func(s string) ([]byte, error) {
		decoded, err := enc.DecodeString(s)
		return decoded, errors.Wrapf(err, "invalid base64 encoded data '%s'", s)
	}, keys...)
}

func URL(fallback string, keys ...string) *url.URL {
	var (
		err    error
		parsed *url.URL
	)

	if parsed, err = url.Parse(fallback); err != nil {
		panic(errors.Wrap(err, "must provide a valid fallback url"))
	}

	return envval(parsed, func(s string) (*url.URL, error) {
		decoded, err := url.Parse(s)
		return decoded, errors.WithStack(err)
	}, keys...)
}

func envval[T any](fallback T, parse func(string) (T, error), keys ...string) T {
	for _, k := range keys {
		s := strings.TrimSpace(os.Getenv(k))
		if s == "" {
			continue
		}

		decoded, err := parse(s)
		if err != nil {
			log.Printf("%s stored an invalid value %v\n", k, err)
			continue
		}

		return decoded
	}

	return fallback
}
