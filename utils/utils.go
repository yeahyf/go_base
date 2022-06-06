package utils

import (
	"io"

	"github.com/yeahyf/go_base/log"
)

func CloseAction(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Errorf("close resource err!", err)
	}
}
