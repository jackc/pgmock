package pgmsg

import (
	"io"
)

type Message interface {
	Encode() ([]byte, error)
	WriteTo(w io.Writer) (int64, error)
}
