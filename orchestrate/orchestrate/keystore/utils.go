package keystore

import (
	"encoding/binary"
	"fmt"
)

func encodeByteSlice(in ...[]byte) []byte {
	l := 0
	for _, v := range in {
		l += len(v)
	}
	if l > 4294967295 {
		panic(fmt.Errorf("input byte slice is too long"))
	}
	out := make([]byte, 4+l)
	binary.BigEndian.PutUint32(out, uint32(l))

	start := 4 + copy(out[4:], in[0])
	if len(in) > 1 {
		for _, v := range in[1:] {
			copy(out[start:], v)
		}
	}
	return out
}
