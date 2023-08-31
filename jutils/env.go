package jutils

import (
	"fmt"
	"os"
)

func LoadEnvVarOrFallback(varId string, fallback string) string {
	value := os.Getenv(varId)
	if len(value) == 0 {
		value = fallback
	}
	return value
}

func LoadEnvVarOrPanic(varId string) string {
	value := os.Getenv(varId)
	if len(value) == 0 {
		panic(fmt.Sprintf("No valid %s", varId))
	}
	return value
}
