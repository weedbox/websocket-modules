package jsonrpc

import (
	"io"
	"reflect"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/weedbox/websocket-modules/websocket_server"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type JSONRPCErrorCode int32

const (
	JSONRPCError_ParseError     JSONRPCErrorCode = -32700
	JSONRPCError_InvalidRequest                  = -32600
	JSONRPCError_NotFound                        = -32601
	JSONRPCError_InvalidParams                   = -32602
	JSONRPCError_InternalError                   = -32603
	JSONRPCError_ServerError                     = -32000
)

var JSONRPCErrorMap = []JSONRPCErrorCode{
	JSONRPCError_InvalidRequest,
	JSONRPCError_NotFound,
	JSONRPCError_InvalidParams,
	JSONRPCError_InternalError,
	JSONRPCError_ServerError,
}

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Result  interface{} `json:"result"`
}

type JSONRPCErrorInfo struct {
	Code    JSONRPCErrorCode `json:"code"`
	Message string           `json:"message"`
	Data    interface{}      `json:"data,omitempty"`
}

type JSONRPCError struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      int64            `json:"id"`
	Error   JSONRPCErrorInfo `json:"error"`
}

type NotificationEntry struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type JSONRPC struct {
}

var rpcRequestPool = sync.Pool{
	New: func() interface{} {
		return &JSONRPCRequest{}
	},
}

var responsePool = sync.Pool{
	New: func() interface{} {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
		}
	},
}

var notificationPool = sync.Pool{
	New: func() interface{} {
		return &NotificationEntry{
			JSONRPC: "2.0",
		}
	},
}

func (je *JSONRPC) ParseRequest(r io.Reader) (*websocket_server.RPCRequest, error) {

	// Allocate request object
	jreq := rpcRequestPool.Get().(*JSONRPCRequest)
	jreq.Params = nil

	// Attempt to decode
	err := json.NewDecoder(r).Decode(jreq)
	if err == io.EOF {

		rpcRequestPool.Put(jreq)

		// One value is expected in the message.
		err = io.ErrUnexpectedEOF

		return nil, err
	}

	if err != nil {
		rpcRequestPool.Put(jreq)
		return nil, err
	}

	// Prepare parameters
	v := reflect.ValueOf(jreq.Params)
	params := make([]interface{}, 0)
	switch v.Kind() {
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		params = jreq.Params.([]interface{})
	default:
		params = append(params, jreq.Params)
	}

	// Create standard request
	req := &websocket_server.RPCRequest{
		ID:     jreq.ID,
		Method: jreq.Method,
		Params: params,
	}

	return req, nil
}

func (je *JSONRPC) PrepareResponse(res *websocket_server.RPCResponse) ([]byte, error) {

	if res.Error != nil {
		return je.createErrorFromObject(res.ID, res.Error)
	}

	// Create response
	data, err := je.createResponse(res.ID, res.Result)
	if err != nil {
		return je.createErrorFromObject(res.ID, err)
	}

	return data, nil
}

func (je *JSONRPC) createErrorFromObject(id int64, err interface{}) ([]byte, error) {

	switch err.(type) {
	case *websocket_server.RPCError:
		rpcError := err.(*websocket_server.RPCError)

		// using customized error code by default
		code := JSONRPCErrorCode(rpcError.Code)
		if JSONRPCErrorCode(len(JSONRPCErrorMap)) > code {
			// Convert to standard JSON-RPC error code
			code = JSONRPCErrorMap[code]
		}

		return je.createError(id, code, rpcError.Message, rpcError.Data)
	default:
		return je.createError(id, JSONRPCError_InternalError, "Internal error", err.(error).Error())
	}
}

func (je *JSONRPC) PrepareNotification(eventName string, payload interface{}) ([]byte, error) {

	// Create or get notification object from pool
	entry := notificationPool.Get().(*NotificationEntry)
	entry.JSONRPC = "2.0"
	entry.Method = eventName
	entry.Params = payload

	// Convert to JSON string
	jsonStr, err := json.Marshal(entry)

	notificationPool.Put(entry)

	if err != nil {
		return []byte(""), err
	}

	return jsonStr, nil
}

func (je *JSONRPC) createResponse(id int64, result interface{}) ([]byte, error) {

	// Create or get response object from pool
	response := responsePool.Get().(*JSONRPCResponse)
	response.JSONRPC = "2.0"
	response.ID = id
	response.Result = result

	// Convert to JSON string
	jsonStr, err := json.Marshal(response)

	responsePool.Put(response)

	if err != nil {
		return []byte(""), err
	}

	return jsonStr, nil
}

func (je *JSONRPC) createError(id int64, code JSONRPCErrorCode, message string, errData interface{}) ([]byte, error) {

	errorEntry := JSONRPCError{
		JSONRPC: "2.0",
		ID:      id,
		Error: JSONRPCErrorInfo{
			Code:    code,
			Message: message,
			Data:    errData,
		},
	}

	jsonStr, err := json.Marshal(errorEntry)
	if err != nil {
		return []byte(""), err
	}

	return jsonStr, nil
}
