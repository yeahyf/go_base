package utils

import (
	"encoding/binary"
	"io"

	"github.com/yeahyf/go_base/log"
)

func CloseAction(c io.Closer) {
	if c != nil {
		err := c.Close()
		if err != nil {
			log.Errorf("close resource err: %v", err)
		}
	}
}

// GetBytesForInt64 将int64转为bytes
func GetBytesForInt64(t uint64) []byte {
	cn := make([]byte, 8)
	binary.LittleEndian.PutUint64(cn, t)
	return cn
}

// GetInt64FromBytes 将bytes转为int64
func GetInt64FromBytes(cn []byte) int64 {
	if cn == nil || len(cn) < 8 {
		return 0
	}
	return int64(binary.LittleEndian.Uint64(cn))
}
