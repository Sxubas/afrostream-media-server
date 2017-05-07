package ts

import (
	"strings"
	"encoding/hex"
)

type Bytes interface {
	ToBytes() (data Data)
	Size() (int)
}

func StrToBytes(s string) (bytes []byte) {
	split := strings.Split(s, " ")
	bytes = make([]byte, len(split))
	for i, part := range split {
		bytesTMP, _ := hex.DecodeString(part)
		bytes[i] =  bytesTMP[0]
	}
	return
}
