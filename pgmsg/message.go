package pgmsg

type Message interface {
	Encode() ([]byte, error)
}

func Parse(typeByte byte, body []byte) (Message, error) {
	switch typeByte {
	case 'R':
		return ParseAuthentication(typeByte, body)
	case 'E':
		return ParseErrorResponse(typeByte, body)
	case 'S':
		return ParseParameterStatus(typeByte, body)
	case 'p':
		return ParsePasswordMessage(typeByte, body)
	default:
		return ParseUnknownMessage(typeByte, body)
	}
}
