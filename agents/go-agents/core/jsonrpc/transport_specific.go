package jsonrpc

// NativeConn provides low level interface for jsonrpc.Channel
// to communicate with native connection such as websocket.
type NativeConn interface {

	// Write writes bytes to the connection.
	// If connection is closed an error of type jsonrpc.CloseError
	// must be returned.
	Write(body []byte) error

	// Next is blocking read of incoming messages.
	// If connection is closed an error of type jsonrpc.CloseError
	// must be returned.
	Next() ([]byte, error)

	// Closes this connection.
	// After it is closed calls to Write and Next must fail with
	// jsonrpc.CloseError error.
	Close() error
}
