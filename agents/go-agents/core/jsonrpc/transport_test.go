package jsonrpc_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/eclipse/che/agents/go-agents/core/jsonrpc"
	"sync"
	"testing"
	"time"
)

// TODO add timeouts to all the wait calls
// TODO close reqRecorder in places where not closed

func TestChannelSaysHello(t *testing.T) {
	beforeConnected := time.Now()

	// initialization routine
	channel, connRecorder, _ := newTestChannel()
	channel.Go()
	defer channel.Close()

	// send hello notification
	channel.SayHello()

	// wait while this notification is received by connection
	connRecorder.WaitUntil(WriteCalledAtLeast(1))

	// check the received notification is expected one
	helloNotification := &jsonrpc.ChannelConnected{}
	connRecorder.UnmarshalRequestParams(0, helloNotification)

	if helloNotification.ChannelID != channel.ID {
		t.Fatalf("Channel ids are different %s != %s", helloNotification.ChannelID, channel.ID)
	}
	if helloNotification.Text != "Hello!" {
		t.Fatalf("Expected text to be 'Hello' but it is %s", helloNotification.Text)
	}
	now := time.Now()
	if !beforeConnected.Before(helloNotification.Time) || !helloNotification.Time.Before(now) {
		t.Fatalf("Expected event time to be between %v < x < %v", beforeConnected, now)
	}
}

// X Notification -> X'
func TestSendingNotification(t *testing.T) {
	channel, connRecorder, _ := newTestChannel()
	channel.Go()
	defer channel.Close()

	method := "event:my-event"
	channel.Notify(method, &testStruct{"Test"})

	connRecorder.WaitUntil(WriteCalledAtLeast(1))

	// check request
	req, err := connRecorder.GetRequest(0)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != method {
		t.Fatalf("Expected to send %s method but sent %s", method, req.Method)
	}
	if !req.IsNotification() {
		t.Fatalf("Expected request to be notification but it has id %v", req.ID)
	}

	// check params
	event := &testStruct{}
	json.Unmarshal(req.RawParams, event)
	if event.Data != "Test" {
		t.Fatal("Expected event data to be 'Test'")
	}
}

// X Request -> X'
func TestSendingRequest(t *testing.T) {
	channel, connRecorder, _ := newTestChannel()
	channel.Go()
	defer channel.Close()

	method := "domain.doSomething"
	channel.Request(method, &testStruct{"Test"}, func(response *jsonrpc.Response, err error) {
		// do nothing
	})

	connRecorder.WaitUntil(WriteCalledAtLeast(1))

	// check request
	req, err := connRecorder.GetRequest(0)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != method {
		t.Fatalf("Expected to send %s method but sent %s", method, req.Method)
	}
	if req.IsNotification() {
		t.Fatal("Expected request not to be notification but it does not have id")
	}

	// check params
	event := &testStruct{}
	json.Unmarshal(req.RawParams, event)
	if event.Data != "Test" {
		t.Fatal("Expected event data to be 'Test'")
	}
}

// X' Request -> X
func TestReceivingRequest(t *testing.T) {
	channel, connRecorder, reqRecorder := newTestChannel()
	channel.Go()
	defer channel.Close()
	defer reqRecorder.Close()

	// prepare a test request object and put it in native connection read stream
	reqBody, err := json.Marshal(testStruct{"Test"})
	if err != nil {
		t.Fatal(err)
	}
	sentReq := &jsonrpc.Request{
		ID:        "1",
		Method:    "domain.doSomething",
		RawParams: reqBody,
	}
	connRecorder.PushNext(sentReq)

	// channel needs some time to call the handler
	reqRecorder.WaitHandleCalled(1)

	receivedReq, _ := reqRecorder.Get(0)
	if receivedReq.Method != sentReq.Method {
		t.Fatalf("Sent method %s but received %s", sentReq.Method, receivedReq.Method)
	}
	if receivedReq.ID != sentReq.ID {
		t.Fatalf("Sent id %T %v but received %T %v", sentReq.ID, sentReq.ID, receivedReq.ID, receivedReq.ID)
	}
	if string(receivedReq.RawParams) != string(sentReq.RawParams) {
		t.Fatalf("Sent params %s but received %s", string(sentReq.RawParams), string(receivedReq.RawParams))
	}
}

// X' Request  -> X
// X' <- Response X
func TestSendingResponseBack(t *testing.T) {
	channel, connRecorder, reqRecorder := newTestChannel()
	channel.Go()
	defer channel.Close()
	defer reqRecorder.Close()

	// prepare a test request object and put it in native connection read stream
	reqBody, err := json.Marshal(testStruct{"Test"})
	if err != nil {
		t.Fatal(err)
	}
	req := &jsonrpc.Request{
		ID:        1,
		Method:    "domain.doSomething",
		RawParams: reqBody,
	}
	connRecorder.PushNext(req)

	// wait for request to arrive
	reqRecorder.WaitHandleCalled(1)

	// respond back
	_, transmitter := reqRecorder.Get(0)
	sentBody := testStruct{"response test data"}
	transmitter.Send(sentBody)

	// wait for response to be written
	connRecorder.WaitUntil(WriteCalledAtLeast(1))

	resp, err := connRecorder.GetResponse(0)
	if err != nil {
		t.Fatal(err)
	}

	// check the response is ok
	if resp.ID != req.ID {
		t.Fatalf("Expected ids to be the same but resp id %v != req id %v", resp.ID, req.ID)
	}
	if resp.Error != nil {
		t.Fatalf("Expected to get response without error, but got %d %s", resp.Error.Code, resp.Error.Message)
	}
	respBody := testStruct{}
	if err := json.Unmarshal(resp.Result, &respBody); err != nil {
		t.Fatal(err)
	}
	if respBody != sentBody {
		t.Fatalf("Expected to get the same body but got %v != %v", respBody, sentBody)
	}
}

// X Request  -> X'
// X <- Response X'
func TestRequestResponseHandling(t *testing.T) {
	channel, connRecorder, reqRecorder := newTestChannel()
	channel.Go()
	defer channel.Close()
	defer reqRecorder.Close()

	respChan := make(chan *jsonrpc.Response, 1)

	// X Request -> X'
	channel.Request("domain.doSomething", &testStruct{"req-params"}, func(response *jsonrpc.Response, err error) {
		respChan <- response
	})

	// wait for the response and catch its id
	connRecorder.WaitUntil(WriteCalledAtLeast(1))
	req, err := connRecorder.GetRequest(0)
	if err != nil {
		t.Fatal(t)
	}

	// X' Response -> X
	repsBody := testStruct{"resp-body"}
	marshaledBody, err := json.Marshal(&repsBody)
	if err != nil {
		t.Fatal(err)
	}
	connRecorder.PushNext(&jsonrpc.Response{
		ID:     req.ID,
		Result: marshaledBody,
	})

	// wait for the response handler function to be called
	select {
	case resp := <-respChan:
		if bytes.Compare(resp.Result, marshaledBody) != 0 {
			t.Fatalf("Received different response body %s != %s", string(resp.Result), string(marshaledBody))
		}
	case <-time.After(time.Second * 2):
		t.Fatal("Didn't receieve the response in 2seconds")
	}
}

// WaitPredicate is used to wait on recorder until the condition
// behind this predicate is met.
type WaitPredicate func(recorder *NativeConnRecorder) bool

func WriteCalledAtLeast(times int) WaitPredicate {
	return func(recorder *NativeConnRecorder) bool {
		return len(recorder.capturedWrites) >= times
	}
}

type testStruct struct {
	Data string `json:"data"`
}

func newTestChannel() (*jsonrpc.Channel, *NativeConnRecorder, *RequestsRecorder) {
	connRecorder := NewNativeConnRecorder()
	reqRecorder := NewRecordingRequestHandler()
	channel := jsonrpc.NewChannel(connRecorder, reqRecorder)
	return channel, connRecorder, reqRecorder
}

type reqPair struct {
	request     *jsonrpc.Request
	transmitter jsonrpc.ResponseTransmitter
}

type RequestsRecorder struct {
	mutex    *sync.Mutex
	cond     *sync.Cond
	closed   bool
	requests []*reqPair
}

func NewRecordingRequestHandler() *RequestsRecorder {
	mx := &sync.Mutex{}
	return &RequestsRecorder{
		mutex:    mx,
		cond:     sync.NewCond(mx),
		closed:   false,
		requests: make([]*reqPair, 0),
	}
}

func (rrh *RequestsRecorder) Handle(r *jsonrpc.Request, rt jsonrpc.ResponseTransmitter) {
	rrh.mutex.Lock()
	defer rrh.mutex.Unlock()
	rrh.requests = append(rrh.requests, &reqPair{request: r, transmitter: rt})
	rrh.cond.Broadcast()
}

func (rrh *RequestsRecorder) WaitHandleCalled(times int) {
	rrh.cond.L.Lock()
	for !rrh.closed && len(rrh.requests) < times {
		rrh.cond.Wait()
	}
	rrh.cond.L.Unlock()
}

func (rrh *RequestsRecorder) Get(idx int) (*jsonrpc.Request, jsonrpc.ResponseTransmitter) {
	rrh.mutex.Lock()
	defer rrh.mutex.Unlock()
	pair := rrh.requests[idx]
	return pair.request, pair.transmitter
}

func (rrh *RequestsRecorder) Close() {
	rrh.mutex.Lock()
	defer rrh.mutex.Unlock()
	rrh.closed = true
	rrh.cond.Broadcast()
}

type NativeConnRecorder struct {
	mutex          *sync.Mutex
	cond           *sync.Cond
	capturedWrites [][]byte
	nextChan       chan []byte
	closed         bool
}

func (ncr *NativeConnRecorder) GetCaptured(idx int) []byte {
	return ncr.capturedWrites[idx]
}

func (ncr *NativeConnRecorder) GetAllCaptured() [][]byte {
	// TODO wrap in lock
	return ncr.capturedWrites
}

func (ncr *NativeConnRecorder) Unmarshal(idx int, v interface{}) error {
	return json.Unmarshal(ncr.GetCaptured(idx), v)
}

func (ncr *NativeConnRecorder) GetRequest(idx int) (*jsonrpc.Request, error) {
	req := &jsonrpc.Request{}
	err := ncr.Unmarshal(idx, req)
	return req, err
}

func (ncr *NativeConnRecorder) GetResponse(idx int) (*jsonrpc.Response, error) {
	resp := &jsonrpc.Response{}
	err := ncr.Unmarshal(idx, resp)
	if floatId, ok := resp.ID.(float64); ok {
		resp.ID = int(floatId)
	}
	return resp, err
}

func (ncr *NativeConnRecorder) UnmarshalRequestParams(idx int, v interface{}) error {
	req, err := ncr.GetRequest(idx)
	if err != nil {
		return err
	}
	err = json.Unmarshal(req.RawParams, &v)
	if err != nil {
		return err
	}
	return nil
}

func (ncr *NativeConnRecorder) WaitUntil(p WaitPredicate) {
	ncr.cond.L.Lock()
	for !ncr.closed && !p(ncr) {
		ncr.cond.Wait()
	}
	ncr.cond.L.Unlock()
}

func (ncr *NativeConnRecorder) PushNext(v interface{}) error {
	marshaled, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ncr.PushNextRaw(marshaled)
	return nil
}

func (ncr *NativeConnRecorder) PushNextRaw(data []byte) {
	ncr.nextChan <- data
}

func NewNativeConnRecorder() *NativeConnRecorder {
	mx := &sync.Mutex{}
	return &NativeConnRecorder{
		mutex:          mx,
		cond:           sync.NewCond(mx),
		capturedWrites: make([][]byte, 0),
		nextChan:       make(chan []byte),
	}
}

func (ncr *NativeConnRecorder) Write(body []byte) error {
	ncr.mutex.Lock()
	ncr.capturedWrites = append(ncr.capturedWrites, body)
	ncr.cond.Broadcast()
	ncr.mutex.Unlock()
	return nil
}

func (ncr *NativeConnRecorder) Next() ([]byte, error) {
	data, ok := <-ncr.nextChan
	if !ok {
		return nil, jsonrpc.NewCloseErr(errors.New("Closed"))
	}
	return data, nil
}

func (ncr *NativeConnRecorder) Close() error {
	ncr.mutex.Lock()
	ncr.closed = true
	ncr.cond.Broadcast()
	ncr.mutex.Unlock()
	close(ncr.nextChan)
	return nil
}
