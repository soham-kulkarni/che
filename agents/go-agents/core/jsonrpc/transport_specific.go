package jsonrpc

type NativeConn interface {
	Write(body []byte) error
	Next() ([]byte, error)
	Close() error
}
