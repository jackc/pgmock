package pgplayback

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgproto3/v2"
)

func (t *Transport) getBackendMessage(data []byte) (msg pgproto3.BackendMessage, err error) {
	var mtype struct {
		Type string
	}
	if err := json.Unmarshal(data, &mtype); err != nil {
		return nil, fmt.Errorf("cannot decipher backend message type: %w", err)
	}

	// unmarshal controls the fields that doesn't require custom marshaling.
	var unmarshal bool
	switch mtype.Type {
	case "AuthenticationCleartextPassword":
		msg = &pgproto3.AuthenticationCleartextPassword{}
	case "AuthenticationMD5Password":
		msg, unmarshal = &pgproto3.AuthenticationMD5Password{}, true
	case "AuthenticationOK":
		msg = &pgproto3.AuthenticationOk{}
	case "AuthenticationSASL":
		msg, unmarshal = &pgproto3.AuthenticationSASL{}, true
	case "AuthenticationSASLContinue":
		msg, unmarshal = &pgproto3.AuthenticationSASLContinue{}, true
	case "AuthenticationSASLFinal":
		msg, unmarshal = &pgproto3.AuthenticationSASLFinal{}, true
	case "BackendKeyData":
		msg, unmarshal = &pgproto3.BackendKeyData{}, true
	case "BindComplete":
		msg = &pgproto3.BindComplete{}
	case "CloseComplete":
		msg = &pgproto3.CloseComplete{}
	case "CommandComplete":
		msg, unmarshal = &pgproto3.CommandComplete{}, true
	case "CopyBothResponse":
		msg, unmarshal = &pgproto3.CopyBothResponse{}, true
	case "CopyData":
		msg, unmarshal = &pgproto3.CopyData{}, true
	case "CopyDone":
		msg = &pgproto3.CopyDone{}
	case "CopyInResponse":
		msg, unmarshal = &pgproto3.CopyInResponse{}, true
	case "CopyOutResponse":
		msg, unmarshal = &pgproto3.CopyOutResponse{}, true
	case "DataRow":
		msg, unmarshal = &pgproto3.DataRow{}, true
	case "": // EmptyQueryResponse
		msg = &pgproto3.EmptyQueryResponse{}
	case "ErrorResponse":
		msg, unmarshal = &pgproto3.ErrorResponse{}, true
	case "FunctionCallResponse":
		msg, unmarshal = &pgproto3.FunctionCallResponse{}, true
	case "NoData":
		msg = &pgproto3.NoData{}
	case "NoticeResponse":
		msg, unmarshal = &pgproto3.NoticeResponse{}, true
	case "NotificationResponse":
		msg, unmarshal = &pgproto3.NotificationResponse{}, true
	case "ParameterDescription":
		msg, unmarshal = &pgproto3.ParameterDescription{}, true
	case "ParameterStatus":
		msg, unmarshal = &pgproto3.ParameterStatus{}, true
	case "ParseComplete":
		msg = &pgproto3.ParseComplete{}
	case "PortalSuspended":
		msg = &pgproto3.PortalSuspended{}
	case "ReadyForQuery":
		msg, unmarshal = &pgproto3.ReadyForQuery{}, true
	case "RowDescription":
		msg, unmarshal = &pgproto3.RowDescription{}, true
	default:
		return nil, fmt.Errorf("unknown backend message type %q", mtype.Type)
	}

	if unmarshal {
		if err := unmarshalJSON(mtype.Type, data, &msg); err != nil {
			return nil, err
		}
	}
	return
}

func (t *Transport) getFrontendMessage(data []byte) (msg pgproto3.FrontendMessage, err error) {
	var mtype struct {
		Type string
	}
	if err := json.Unmarshal(data, &mtype); err != nil {
		return nil, fmt.Errorf("cannot decipher frontend message type: %w", err)
	}

	var unmarshal bool
	switch mtype.Type {
	case "Bind":
		msg, unmarshal = &pgproto3.Bind{}, true
	case "CancelRequest":
		msg, unmarshal = &pgproto3.CancelRequest{}, true
	case "Close":
		msg, unmarshal = &pgproto3.Close{}, true
	case "CopyData":
		msg, unmarshal = &pgproto3.CopyData{}, true
	case "CopyDone":
		msg = &pgproto3.CopyDone{}
	case "CopyFail":
		msg, unmarshal = &pgproto3.CopyFail{}, true
	case "Describe":
		msg, unmarshal = &pgproto3.Describe{}, true
	case "Execute":
		msg, unmarshal = &pgproto3.Execute{}, true
	case "Flush":
		msg = &pgproto3.Flush{}
	case "GSSEncRequest":
		msg = &pgproto3.GSSEncRequest{}
	case "Parse":
		msg, unmarshal = &pgproto3.Parse{}, true
	case "PasswordMessage":
		// TODO: Probably don't unmarshal to don't reveal the password. Maybe add as an option.
		msg, unmarshal = &pgproto3.PasswordMessage{}, true
	case "Query":
		msg, unmarshal = &pgproto3.Query{}, true
	case "SASLInitialResponse":
		msg, unmarshal = &pgproto3.SASLInitialResponse{}, true
	case "SASLResponse":
		msg, unmarshal = &pgproto3.SASLResponse{}, true
	case "SSLRequest":
		msg = &pgproto3.SSLRequest{}
	case "StartupMessage":
		msg, unmarshal = &pgproto3.StartupMessage{}, true
	case "Sync":
		msg = &pgproto3.Sync{}
	case "Terminate":
		msg = &pgproto3.Terminate{}
	default:
		return nil, fmt.Errorf("unknown frontend message type %q", mtype.Type)
	}

	if unmarshal {
		if err := unmarshalJSON(mtype.Type, data, &msg); err != nil {
			return nil, err
		}
	}
	return
}

func unmarshalJSON(msgType string, data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("cannot process message %q: %w", msgType, err)
	}
	return nil
}
