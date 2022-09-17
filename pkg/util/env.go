package util

import (
	"os"
	"strconv"
	"strings"
)

// EnvStr returns the value of the environment variable named by the key.
func EnvStr(key, defaultValue string) string {
	if value, exist := os.LookupEnv(key); exist {
		return value
	}

	return defaultValue
}

// EnvInt returns the value of the environment variable named by the key.
func EnvInt(key string, defaultValue int) int {
	if value, exist := os.LookupEnv(key); exist {
		number, err := strconv.Atoi(value)
		if err != nil {
			// I think that we should log this error
			return defaultValue
		}

		return number
	}

	return defaultValue
}

// EnvArr returns the value of the environment variable named by the key.
func EnvArr(key string, defaultValue []string) []string {
	if value, exist := os.LookupEnv(key); exist {
		return strings.Split(value, ",")
	}

	return defaultValue
}
