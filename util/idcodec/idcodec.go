package idcodec

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/sqids/sqids-go"
)

const (
	minLength       = 6
	defaultAlphabet = "k2mYLSvFpJBqNgdWfXrOa3bjRhVnQHx8DzT6UsEC7wGe9KZc4A5t"
)

var (
	instance *sqids.Sqids
	once     sync.Once
)

func codec() *sqids.Sqids {
	once.Do(func() {
		alphabet := defaultAlphabet
		if env := os.Getenv("IDCODEC_ALPHABET"); env != "" {
			alphabet = env
		}

		var err error
		instance, err = sqids.New(sqids.Options{
			Alphabet:  alphabet,
			MinLength: minLength,
		})
		if err != nil {
			panic(fmt.Sprintf("idcodec: failed to initialise sqids with alphabet %q: %v", alphabet, err))
		}
	})
	return instance
}

func Encode(id int64) (string, error) {
	if id < 0 {
		return "", fmt.Errorf("idcodec: cannot encode negative id %d", id)
	}
	encoded, err := codec().Encode([]uint64{uint64(id)})
	if err != nil {
		return "", fmt.Errorf("idcodec: encode failed for id %d: %w", id, err)
	}
	return encoded, nil
}

func MustEncode(id int64) string {
	if id == 0 {
		return ""
	}
	s, err := Encode(id)
	if err != nil {
		return ""
	}
	return s
}

func Decode(encoded string) (int64, error) {
	if encoded == "" {
		return 0, fmt.Errorf("idcodec: cannot decode empty string")
	}

	if isNumeric(encoded) {
		return 0, fmt.Errorf("idcodec: raw numeric IDs are not accepted: %q", encoded)
	}

	numbers := codec().Decode(encoded)
	if len(numbers) == 0 {
		return 0, fmt.Errorf("idcodec: decode produced no values for %q", encoded)
	}

	reEncoded, err := codec().Encode(numbers)
	if err != nil || reEncoded != encoded {
		return 0, fmt.Errorf("idcodec: integrity check failed for %q", encoded)
	}

	return int64(numbers[0]), nil
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func IsIDField(fieldName string) bool {
	return fieldName == "id" || fieldName == "Id" || fieldName == "ID" ||
		strings.HasSuffix(fieldName, "_id") || strings.HasSuffix(fieldName, "Id") || strings.HasSuffix(fieldName, "ID")
}
