package util

import (
	"os"
	"strconv"
)

func GetEnv[T any](env string, alt T) T {
	var result interface{}

	// Read the environment variable
	value := os.Getenv(env)
	if value == "" {
		// Return the default value if the environment variable is empty
		return alt
	}

	// Determine the type of the target variable to parse
	switch any(alt).(type) {
	case int:
		val, err := strconv.Atoi(value)
		if err == nil {
			result = val
		}
	case int8:
		val, err := strconv.ParseInt(value, 10, 8)
		if err == nil {
			result = int8(val)
		}
	case int16:
		val, err := strconv.ParseInt(value, 10, 16)
		if err == nil {
			result = int16(val)
		}
	case int32:
		val, err := strconv.ParseInt(value, 10, 32)
		if err == nil {
			result = int32(val)
		}
	case int64:
		val, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			result = int64(val)
		}
	case uint:
		val, err := strconv.ParseUint(value, 10, 0)
		if err == nil {
			result = uint(val)
		}
	case uint8:
		val, err := strconv.ParseUint(value, 10, 8)
		if err == nil {
			result = uint8(val)
		}
	case uint16:
		val, err := strconv.ParseUint(value, 10, 16)
		if err == nil {
			result = uint16(val)
		}
	case uint32:
		val, err := strconv.ParseUint(value, 10, 32)
		if err == nil {
			result = uint32(val)
		}
	case uint64:
		val, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			result = uint64(val)
		}
	case uintptr:
		val, err := strconv.ParseUint(value, 10, 0)
		if err == nil {
			result = uintptr(val)
		}
	case float32:
		val, err := strconv.ParseFloat(value, 32)
		if err == nil {
			result = float32(val)
		}
	case float64:
		val, err := strconv.ParseFloat(value, 64)
		if err == nil {
			result = float64(val)
		}
	case complex64:
		val, err := strconv.ParseComplex(value, 64)
		if err == nil {
			result = complex64(val)
		}
	case complex128:
		val, err := strconv.ParseComplex(value, 128)
		if err == nil {
			result = complex128(val)
		}
	case bool:
		val, err := strconv.ParseBool(value)
		if err == nil {
			result = val
		}
	case string:
		result = value
	default:
		// Return the default value if the type is not supported
		return alt
	}

	// Convert the result to type T and return
	if result == nil {
		return alt
	}
	return result.(T)
}
