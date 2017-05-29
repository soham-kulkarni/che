// Package RPC provides full implementation of JSONRPC 2.0 protocol.
// See http://www.jsonrpc.org/specification.
//
// The implementation specific terminology:
//
//   Transfer channel - a transport layer for two endpoints
//   for bidirectional communication.
//
//   Server - the endpoint of the transfer channel which receives
//   Requests or Notifications and replies with Responses.
//
//   Client - the endpoint of the transfer channel which receives
//   Responses and sends Requests and Notifications.
//
package jsonrpc

import "encoding/json"

const (
	// ParseErrorCode indicates that invalid JSON was received by the server.
	ParseErrorCode = -32700

	// InvalidRequestErrorCode indicates that request object is not valid,
	// fails when route decoder can't decode params.
	InvalidRequestErrorCode = -32600

	// MethodNotFoundErrorCode indicates that there is no route for such method.
	MethodNotFoundErrorCode = -32601

	// InvalidParamsErrorCode indicates that handler parameters are considered as not valid.
	// This error type should be returned directly from the HandlerFunc
	InvalidParamsErrorCode = -32602

	// InternalErrorCode is returned when error returned from the Route HandlerFunc is different from Error type.
	InternalErrorCode = -32603

	// -32000 to -32099 Reserved for implementation-defined server-errors.
)

// Request is the identified call of the method.
// Server MUST eventually reply on the response and include
// the same identifier value as the request provides.
//
// Request without ID is Notification.
// Server MUST NOT reply to Notification.
type Request struct {

	// Version of this request e.g. '2.0'.
	//
	// The version field is required.
	Version string `json:"jsonrpc"`

	// Method is the name which will be proceeded by this request.
	//
	// Must not start with "rpc" + (U+002E or ASCII 46), such methods are
	// reserved for rpc internal methods and extensions.
	//
	// The method field is required.
	Method string `json:"method"`

	// The unique identifier of this operation request.
	// If a client needs to identify the result of the operation execution,
	// the id should be passed by the client, then it is guaranteed
	// that the client will receive the result frame with the same id.
	// The uniqueness of the identifier must be controlled by the client,
	// if client doesn't specify the identifier in the operation call,
	// the response won't contain the identifier as well.
	//
	// It is preferable to specify identifier for those calls which may
	// either validate data, or produce such information which can't be
	// identified by itself.
	//
	// If id is set then the object is Request otherwise it's Notification.
	ID interface{} `json:"id"`

	// Request data, parameters which are needed for operation execution.
	// Params are either json array or json object, for json objects
	// names of the parameters are case sensitive.
	//
	// The params field is optional.
	RawParams json.RawMessage `json:"params"`
}

// Response is a reply on a certain request, which represents the result
// of the certain operation execution.
// Response MUST provide the same identifier as the request which forced it.
type Response struct {

	// Version of this response e.g. '2.0'.
	// The version is required.
	Version string `json:"jsonrpc"`

	// The operation call identifier, will be set only
	// if the operation contains it.
	ID interface{} `json:"id"`

	// Result is the result of the method call.
	// Result can be anything determined by the operation(method).
	// Result and Error are mutually exclusive.
	Result interface{} `json:"result,omitempty"`

	// Result and Error are mutually exclusive.
	// Present only if the operation execution fails due to an error.
	Error *Error `json:"error,omitempty"`
}

// Error indicates any exceptional situation during operation execution,
// e.g an attempt to perform operation using invalid data.
type Error struct {
	error `json:"-"`

	// Code is the value indicating the certain error type.
	Code int `json:"code"`

	// Message is the description of this error.
	Message string `json:"message"`

	// Data any kind of data which provides additional
	// information about the error e.g. stack trace, error time.
	Data interface{} `json:"data,omitempty"`
}
