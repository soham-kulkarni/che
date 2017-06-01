// Package provides implementation of jsonrpc.NativeConn based on websocket.
//
// The example:
//
// Client:
//	conn, err := jsonrpcws.Dial("ws://host:port/path")
//	if err != nil {
//      	panic(err)
//      }
// 	channel := jsonrpc.NewChannel(conn, jsonrpc.DefaultRouter)
//	channel.Go()
//	channel.SayHello()
//
// Server:
//	conn, err := jsonrpcws.Upgrade(w, r)
//	if err != nil {
//      	panic(err)
//      }
//	channel := jsonrpc.NewChannel(conn, jsonrpc.DefaultRouter)
//	channel.Go()
//	channel.SayHello()
package jsonrpcws

import (
	"github.com/eclipse/che-lib/websocket"
	"net/http"
)

var (
	defaultUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func NewDial(url string) (*NativeConnAdapter, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &NativeConnAdapter{conn}, nil
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*NativeConnAdapter, error) {
	conn, err := defaultUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &NativeConnAdapter{conn}, nil
}

type NativeConnAdapter struct {
	conn *websocket.Conn
}

// jsonrpc.NativeConn implementation
func (adapter *NativeConnAdapter) Write(data []byte) error {
	w, err := adapter.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

func (adapter *NativeConnAdapter) Next() ([]byte, error) {
	_, data, err := adapter.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return data, err
}

func (adapter *NativeConnAdapter) Close() error {
	err := adapter.conn.Close()
	if closeErr, ok := err.(*websocket.CloseError); ok && isNormallyClosed(closeErr.Code) {
		return nil
	} else {
		return err
	}
}

func isNormallyClosed(code int) bool {
	return code == websocket.CloseGoingAway ||
		code == websocket.CloseNormalClosure ||
		code == websocket.CloseNoStatusReceived
}
