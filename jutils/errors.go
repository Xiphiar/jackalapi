package jutils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func ProcessError(block string, caughtError error) {
	fmt.Printf("***** Error in block: %s *****\n", block)
	fmt.Println(caughtError)
	fmt.Println("***** End Error Report *****")
}

func ProcessHttpError(block string, caughtError error, eCode int, w http.ResponseWriter) {
	fmt.Printf("***** Error in block: %s *****\n", block)
	fmt.Println(caughtError)
	fmt.Println("***** End Error Report *****")
	w.WriteHeader(eCode)
	_, err := w.Write([]byte(caughtError.Error()))
	if err != nil {
		ProcessError(fmt.Sprintf("processHttpPostError for %s", block), err)
	}
}

func ProcessCustomHttpError(block string, customError string, eCode int, w http.ResponseWriter) error {
	asError := errors.New(strings.ToLower(customError))
	ProcessHttpError(block, asError, eCode, w)
	return asError
}
