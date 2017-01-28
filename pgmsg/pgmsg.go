package pgmsg

import (
	"fmt"
)

type Message interface {
	Encode() ([]byte, error)
}

func Parse(typeByte byte, body []byte) (Message, error) {
	switch typeByte {
	case 'E':
		return ParseErrorResponse(typeByte, body)
	case 'K':
		return ParseBackendKeyData(typeByte, body)
	case 'p':
		return ParsePasswordMessage(typeByte, body)
	case 'R':
		return ParseAuthentication(typeByte, body)
	case 'S':
		return ParseParameterStatus(typeByte, body)
	default:
		return ParseUnknownMessage(typeByte, body)
	}
}

type invalidMessageLenErr struct {
	messageType string
	expectedLen int
	actualLen   int
}

func (e *invalidMessageLenErr) Error() string {
	return fmt.Sprintf("%s body must have length of %d, but it is %d", e.messageType, e.expectedLen, e.actualLen)
}
