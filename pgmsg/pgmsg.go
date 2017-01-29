package pgmsg

import (
	"fmt"
)

type Message interface {
	Encode() ([]byte, error)
}

type FrontendMessage interface {
	Encode() ([]byte, error)
	Frontend() // no-op method to distinguish frontend from backend methods
}

type BackendMessage interface {
	Encode() ([]byte, error)
	Backend() // no-op method to distinguish frontend from backend methods
}

func ParseBackend(typeByte byte, body []byte) (BackendMessage, error) {
	switch typeByte {
	case 'E':
		return ParseErrorResponse(body)
	case 'K':
		return ParseBackendKeyData(body)
	case 'R':
		return ParseAuthentication(body)
	case 'S':
		return ParseParameterStatus(body)
	case 'T':
		return ParseRowDescription(body)
	case 'Z':
		return ParseReadyForQuery(body)
	default:
		return ParseUnknownMessage(typeByte, body)
	}
}

func ParseFrontend(typeByte byte, body []byte) (FrontendMessage, error) {
	switch typeByte {
	case 'p':
		return ParsePasswordMessage(body)
	case 'Q':
		return ParseQuery(body)
	case 'X':
		return ParseTerminate(body)
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

type invalidMessageFormatErr struct {
	messageType string
}

func (e *invalidMessageFormatErr) Error() string {
	return fmt.Sprintf("%s body is invalid", e.messageType)
}
