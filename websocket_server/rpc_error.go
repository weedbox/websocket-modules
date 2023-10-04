package websocket_server

type RPCErrorCode int32

const (
	ErrorCode_InvalidRequest                       RPCErrorCode = 1000
	ErrorCode_NotFound                                          = 2000
	ErrorCode_InvalidParams                                     = 3000
	ErrorCode_InvalidParams_Invalid_Arguments                   = 3001
	ErrorCode_InvalidParams_Insufficient_Arguments              = 3002
	ErrorCode_InternalError                                     = 4000
	ErrorCode_ServerError                                       = 5000
)

var (
	errorMsgMap = map[RPCErrorCode]string{
		ErrorCode_InvalidRequest:                       "Invalid Request",
		ErrorCode_NotFound:                             "Method not found",
		ErrorCode_InvalidParams:                        "Invalid params",
		ErrorCode_InvalidParams_Invalid_Arguments:      "Invalid arguments",
		ErrorCode_InvalidParams_Insufficient_Arguments: "Insufficient arguments",
		ErrorCode_InternalError:                        "Internal error",
		ErrorCode_ServerError:                          "Server error",
	}
)

type RPCError struct {
	Code    RPCErrorCode `json:"code"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return e.Message
}

func NewError(code RPCErrorCode, data interface{}) *RPCError {
	return &RPCError{
		Code:    code,
		Message: errorMsgMap[code],
		Data:    data,
	}
}
