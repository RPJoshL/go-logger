package logger

import (
	"os"
	"strings"
)

// getEnvBool returns a boolean value read from the environmnet variable.
// If the variable was not set, the default value will be returned
func getEnvBool(name string, defaultValue bool) bool {
	val := defaultValue
	if strVal, isSet := os.LookupEnv(name); isSet {
		strVal = strings.ToLower(strVal)
		return strVal == "1" || strVal == "true" || strVal == "yes" || strVal == "ja"
	}

	return val
}

// getEnvString returns a string value read from the environmnet variable.
// If the variable was not set, the default value will be returned
func getEnvString(name string, defaultValue string) string {
	val := defaultValue
	if strVal, isSet := os.LookupEnv(name); isSet {
		val = strVal
	}

	return val
}
