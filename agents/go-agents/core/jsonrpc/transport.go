package jsonrpc

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// TODO response transmitter timeout
// TODO q.watch()
// TODO develop approach for unmarshalling response error data field

const (
	DefaultVersion                  = "2.0"
	ConnectedNotificationMethodName = "connected"
)

var (
	prevChanID    uint64
	prevRequestID int64
)

// Channel is high level jsonrpc transport layer which
// uses native connection to access low level transport routines.
type Channel struct {

	// ID is the unique identifier of this channel.
	ID string

	// Created is the time when this channel was created.
	Created time.Time

	conn        NativeConn
	jsonOutChan chan interface{}
	reqHandler  RequestHandler
	rq          *requestQ
}

// ResponseHandlerFunc used to handle requests responses.
// The request is sent by one of the Request, RequestBare, RequestRaw methods.
// If the response doesn't arrive in time the handler func will be called
// with an error of type TimeoutError.
// Note 'response.Error' has nothing to do with func err param.
type ResponseHandlerFunc func(response *Response, err error)

// RequestHandler is a single handler for all the channel incoming requests.
// The goal of the handler is to dispatch each incoming request to the interested side
// e.g. jsonrpc.Router dispatches registered request to the routes.
// If the request is not handled in a configured period of time, then
// the channel will automatically send error response back and the following
// call to ResponseTransmitter will be ignored.
type RequestHandler interface {
	Handle(request *Request, rt ResponseTransmitter)
}

// ResponseTransmitter provides interface which allows to respond to request.
// The implementation must guarantee that reply will be eventually sent.
// Functions Send & SendError MUST not be called both or twice on the same
// instance of transmitter.
// The request id MUST be included to the response.
type ResponseTransmitter interface {

	// Send sends jsonrpc response with a given result in body.
	// This function can be called only once on one transmitter instance.
	Send(result interface{})

	// SendError sends jsonrpc response with a given error in body.
	// This function can be called only once on one transmitter instance.
	SendError(err Error)

	// Channel returns the channel this transmitter belongs to.
	Channel() Channel
}

// CloseError is an error which MUST be
// published by NativeConn implementations and used to determine
// the cases when channel job should be stopped.
type CloseError struct{ error }

// TimeoutError occurs when timeout is reached before normal handling completes.
type TimeoutError struct{ error }

// NewChannel creates a new channel.
// Use channel.Go() to start the channel.
func NewChannel(conn NativeConn, handler RequestHandler) *Channel {
	return &Channel{
		ID:          "channel-" + strconv.Itoa(int(atomic.AddUint64(&prevChanID, 1))),
		Created:     time.Now(),
		conn:        conn,
		jsonOutChan: make(chan interface{}),
		reqHandler:  handler,
		rq:          &requestQ{pairs: make(map[int64]*rqPair)},
	}
}

// Go starts this channel, makes it functional.
func (c *Channel) Go() {
	go c.mainWriteLoop()
	go c.mainReadLoop()
	go c.rq.watch()
}

// Notify sends notification(request without id) using given params as its body.
// Use NotifyBare in preference to Notify with nil params.
func (c *Channel) Notify(method string, params interface{}) {
	if params == nil {
		c.NotifyBare(method)
	} else {
		if bytes, err := json.Marshal(params); err != nil {
			log.Printf("Could not unmrashall non-nil notification params, it won't be send. Error %s", err.Error())
		} else {
			c.NotifyRaw(method, bytes)
		}
	}
}

// NotifyBare sends notification like Notify does but
// sends no request parameters in it.
func (c *Channel) NotifyBare(method string) {
	c.jsonOutChan <- &Request{Version: DefaultVersion, Method: method}
}

// NotifyRaw sends notification like Notify does using
// given marshaledParams as request params.
// NotifyBare && Notify should be used in preference to this func.
func (c *Channel) NotifyRaw(method string, marshaledParams []byte) {
	c.jsonOutChan <- &Request{
		Version:   DefaultVersion,
		Method:    method,
		RawParams: marshaledParams,
	}
}

// Request sends request marshalling a given params as its body.
// ResponseHandlerFunc will be called as soon as the response arrives,
// or response arrival timeout reached, in that case error of type
// TimeoutError will be passed to the handler.
// Use RequestBare in preference to Request with nil params.
func (c *Channel) Request(method string, params interface{}, rhf ResponseHandlerFunc) {
	if params == nil {
		c.RequestBare(method, rhf)
	} else {
		if bytes, err := json.Marshal(params); err != nil {
			log.Printf("Could not unmrashall non-nil request params, it won't be send. Error %s", err.Error())
		} else {
			c.RequestRaw(method, bytes, rhf)
		}
	}
}

// RequestBare sends the request like Request func does
// but sends no params in it.
func (c *Channel) RequestBare(method string, rhf ResponseHandlerFunc) {
	id := atomic.AddInt64(&prevRequestID, 1)
	request := &Request{ID: id, Method: method}
	c.rq.add(id, request, rhf)
	c.jsonOutChan <- request
}

// RequestRaw sends the request like Request func does
// but using marshaledParams as request params.
// Request & RequestBare should be used in preference to this func.
func (c *Channel) RequestRaw(method string, marshaledParams []byte, rhf ResponseHandlerFunc) {
	id := atomic.AddInt64(&prevRequestID, 1)
	request := &Request{
		ID:        id,
		Method:    method,
		RawParams: marshaledParams,
	}
	c.rq.add(id, request, rhf)
	c.jsonOutChan <- request
}

// Close closes native connection and internal sources, so started
// goroutines should be eventually stopped.
func (c *Channel) Close() {
	// stop must stop all the goroutines started by Channel.Go func.
	// first close out channel so mainWriteLoop stopped
	close(c.jsonOutChan)
	// close native connection which stops mainReadLoop
	c.conn.Close()
	// stop request queue watching routine
	c.rq.stopWatching()
}

// SayHello sends hello notification.
func (c *Channel) SayHello() {
	c.Notify(ConnectedNotificationMethodName, &ChannelConnected{
		Time:      c.Created,
		ChannelID: c.ID,
		Text:      "Hello!",
	})
}

// ChannelConnected struct describing notification params sent by SayHello.
type ChannelConnected struct {
	Time      time.Time `json:"time"`
	ChannelID string    `json:"channel"`
	Text      string    `json:"text"`
}

func (c *Channel) mainWriteLoop() {
	for message := range c.jsonOutChan {
		if bytes, err := json.Marshal(message); err != nil {
			log.Printf("Couldn't marshal message: %T, %v to json. Error %s", message, message, err.Error())
		} else {
			if err := c.conn.Write(bytes); err != nil {
				log.Printf("Couldn't write message to the channel. Message: %T, %v", message, message)
			}
		}
	}
}

func (c *Channel) mainReadLoop() {
	for {
		bytes, err := c.conn.Next()
		if err != nil {
			if _, ok := err.(*CloseError); !ok {
				log.Printf("Couldn't read next message due to occurred error %s", err.Error())

			}
			return
		}

		draft := &draft{}
		err = json.Unmarshal(bytes, draft)
		if err != nil {
			log.Printf("Couldn't unmarshall received frame to request/response. Err %s", err.Error())
		}

		if draft.Method == "" {
			c.dispatchResponse(&Response{
				Version: draft.Version,
				ID:      draft.ID,
				Result:  draft.Result,
				Error:   draft.Error,
			})
		} else {
			c.dispatchRequest(&Request{
				Version:   draft.Version,
				Method:    draft.Method,
				ID:        draft.ID,
				RawParams: draft.RawParams,
			})
		}
	}
}

func (c *Channel) dispatchResponse(r *Response) {
	if r.ID == nil {
		log.Print("Received response with empty identifier, response will be ignored")
		return
	}

	// float64 used for json numbers https://blog.golang.org/json-and-go
	floatID, ok := r.ID.(float64)
	if !ok {
		log.Printf("Received response with non-numeric identifier %T %v, "+
			"response will be ignored", r.ID, r.ID)
		return
	}

	id := int64(floatID)
	rqPair, ok := c.rq.remove(id)
	if ok {
		rqPair.respHandlerFunc(r, nil)
	} else {
		log.Printf("Response handler for request id '%v' is missing which means that response "+
			"arrived to late, or response provides a wrong id", id)
	}
}

func (c *Channel) dispatchRequest(r *Request) {
	if r.IsNotification() {
		c.reqHandler.Handle(r, &notificationRespTransmitter{r})
	} else {
		c.reqHandler.Handle(r, &requestRespTransmitter{
			reqId:   r.ID,
			channel: c,
		})
	}
}

// both request and response
type draft struct {
	Version   string          `json:"jsonrpc"`
	Method    string          `json:"method"`
	ID        interface{}     `json:"id"`
	RawParams json.RawMessage `json:"params"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     *Error          `json:"error,omitempty"`
}

func NewCloseErr(err error) *CloseError {
	return &CloseError{error: err}
}

type rqPair struct {
	request         *Request
	respHandlerFunc ResponseHandlerFunc
	saved           time.Time
}

// Request queue is used for internal request + response handler storage,
// which allows to handle future responses.
// The q does not support any request identifiers except int64.
type requestQ struct {
	sync.RWMutex
	pairs                map[int64]*rqPair
	ticker               *time.Ticker
	stop                 chan bool
	allowedResponseDelay time.Duration
}

func (q *requestQ) watch() {
	//for {
	//	select {
	//	case <-q.ticker.C:
	//		cleanTime := time.Now().Add(-q.allowedResponseDelay)
	//		hermits := make([]int64, 0)
	//		q.RLock()
	//		for id, pair := range q.pairs {
	//			if pair.saved.Before(cleanTime) {
	//				hermits = append(hermits, id)
	//			}
	//		}
	//		q.RUnlock()
	//
	//		for _, id := range hermits {
	//			if rqPair, ok := q.remove(id); ok {
	//				rqPair.respHandlerFunc(nil, &TimeoutError{errors.New("Response didn't arrive in time")})
	//			}
	//		}
	//	case <-q.stop:
	//		return
	//	}
	//}
}

func (q *requestQ) stopWatching() {
	//q.stop <- true
}

func (q *requestQ) add(id int64, r *Request, rhf ResponseHandlerFunc) {
	q.Lock()
	defer q.Unlock()
	q.pairs[id] = &rqPair{
		request:         r,
		respHandlerFunc: rhf,
		saved:           time.Now(),
	}
}

func (q *requestQ) remove(id int64) (*rqPair, bool) {
	q.Lock()
	defer q.Unlock()
	pair, ok := q.pairs[id]
	if ok {
		delete(q.pairs, id)
	}
	return pair, ok
}

type requestRespTransmitter struct {
	reqId   interface{}
	channel *Channel
}

func (drt *requestRespTransmitter) Send(result interface{}) {
	bytes, _ := json.Marshal(result)
	drt.channel.jsonOutChan <- &Response{
		Version: DefaultVersion,
		ID:      drt.reqId,
		Result:  bytes,
	}
}

func (drt *requestRespTransmitter) SendError(err Error) {
	drt.channel.jsonOutChan <- &Response{
		Version: DefaultVersion,
		ID:      drt.reqId,
		Error:   &err,
	}
}

func (drt *requestRespTransmitter) Channel() Channel { return *drt.channel }

type notificationRespTransmitter struct {
	request *Request
}

func (nrt *notificationRespTransmitter) logNoResponse(res interface{}) {
	log.Printf(
		"The response to the notification '%s' will not be send(jsonrpc2.0 spec). The response was %T, %v",
		nrt.request.Method,
		res,
		res,
	)
}

func (nrt *notificationRespTransmitter) Send(result interface{}) { nrt.logNoResponse(result) }

func (nrt *notificationRespTransmitter) SendError(err Error) { nrt.logNoResponse(err) }

func (nrt *notificationRespTransmitter) Channel() Channel { return Channel{} }
