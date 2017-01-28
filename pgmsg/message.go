package pgmsg

type Message interface {
	Encode() ([]byte, error)
}
