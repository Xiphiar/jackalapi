package ipfs

import (
	"bytes"
	"fmt"

	record "github.com/libp2p/go-libp2p-record"
)

type customValidator struct {
	Base record.Validator
}

func (cv customValidator) Validate(key string, value []byte) error {
	fmt.Printf("DHT Validating: %s = %s\n", key, value)
	return cv.Base.Validate(key, value)
}

func (cv customValidator) Select(key string, values [][]byte) (int, error) {
	fmt.Printf("DHT Selecting Among: %s = %s\n", key, bytes.Join(values, []byte("; ")))
	return cv.Base.Select(key, values)
}
