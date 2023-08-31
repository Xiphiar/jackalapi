package jutils

import (
	"bytes"
)

func CloneBytes(reader *bytes.Reader) []byte {
	var allBytes []byte
	reader.Read(allBytes)
	reader.Seek(0, 0)
	return allBytes
}
