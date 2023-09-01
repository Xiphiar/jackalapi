package jutils

import (
	"bytes"
)

func CloneBytes(reader *bytes.Reader) []byte {
	var allBytes []byte
	_, err := reader.Read(allBytes)
	if err != nil {
		return nil
	}
	_, err = reader.Seek(0, 0)
	if err != nil {
		return nil
	}
	return allBytes
}
