package jsonrpc

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// TODO response transmitter timeout
// TODO rq.watch()
// TODO connect to the client
// TODO restrict ids to be only ints and strings

const (
	DefaultVersion                  = "2.0"
	ConnectedNotificationMethodName = "connected"
)

var (
	prevChanID    uint64
	prevRequestID uint64
)

func NewChannel(conn NativeConn, handler RequestHandler) *Channel {
	channel := Channel{}
	channel.ID = "channel-" + strconv.Itoa(int(atomic.AddUint64(&prevChanID, 1)))
	channel.Created = time.Now()
	channel.conn = conn
	channel.jsonOutChan = make(chan interface{})
	channel.prevRequestId = 0
	channel.reqHandler = handler
	channel.rq = &requestQ{pairs: make(map[interface{}]*rqPair)}
	return &channel
}

type ResponseHandlerFunc func(response *Response, err error)

type Channel struct {
	ID      string
	Created time.Time

	conn          NativeConn
	jsonOutChan   chan interface{}
	dropChan      chan bool
	prevRequestId uint64
	reqHandler    RequestHandler
	rq            *requestQ
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
			if _, ok := err.(*CloseErr); !ok {
				log.Printf("Couldn't read next message due to occurred error %s", err.Error())

			}
			return
		}

		draft := &draft{}
		err = json.Unmarshal(bytes, draft)
		if err != nil {
			log.Printf("Couldn't unmarshall received frame to request/response. Err %s", err.Error())
		}

		// TODO add more assertions allow only text and int ids
		// float64 id may be marshaled to id field while int is necessary
		if floatId, ok := draft.ID.(float64); ok {
			draft.ID = int(floatId)
		}

		if draft.Method == "" {
			c.dispatchResp(&Response{
				Version: draft.Version,
				ID:      draft.ID,
				Result:  draft.Result,
				Error:   draft.Error,
			})
		} else {
			c.dispatchReq(&Request{
				Version:   draft.Version,
				Method:    draft.Method,
				ID:        draft.ID,
				RawParams: draft.RawParams,
			})
		}
	}
}

func (c *Channel) dispatchReq(r *Request) {
	if r.IsNotification() {
		c.reqHandler.Handle(r, &notificationRespTransmitter{r})
	} else {
		c.reqHandler.Handle(r, &requestRespTransmitter{
			reqId:   r.ID,
			channel: c,
		})
	}
}

func (c *Channel) dispatchResp(r *Response) {
	rqPair, ok := c.rq.remove(r.ID)
	if !ok {
		log.Printf("Response handler for request id '%s' is missing which means that response "+
			"arrived to late, or response provides a wrong id", r.ID)
	} else {
		rqPair.respHandlerFunc(r, nil)
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

type CloseErr struct {
	error
}

func NewCloseErr(err error) *CloseErr {
	return &CloseErr{error: err}
}

func (c *Channel) Close() {
	// stop must stop all the goroutines started by Channel.Go func.
	// first close out channel so mainWriteLoop stopped
	close(c.jsonOutChan)
	// close native connection which stops mainReadLoop
	c.conn.Close()
	// stop request queue watching routine
	c.rq.stopWatching()
}

func (c *Channel) Go() {
	go c.mainWriteLoop()
	go c.mainReadLoop()
	go c.rq.watch()
}

func (c *Channel) SayHello() {
	c.Notify(ConnectedNotificationMethodName, &ChannelConnected{
		Time:      c.Created,
		ChannelID: c.ID,
		Text:      "Hello!",
	})
}

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

func (c *Channel) NotifyBare(method string) {
	c.jsonOutChan <- &Request{Version: DefaultVersion, Method: method}
}

func (c *Channel) NotifyRaw(method string, params []byte) {
	c.jsonOutChan <- &Request{
		Version:   DefaultVersion,
		Method:    method,
		RawParams: params,
	}
}

func (c *Channel) RequestBare(method string, rhf ResponseHandlerFunc) {
	request := &Request{ID: atomic.AddUint64(&prevRequestID, 1), Method: method}
	c.rq.add(request, rhf)
	c.jsonOutChan <- request
}

func (c *Channel) RequestRaw(method string, params []byte, rhf ResponseHandlerFunc) {
	request := &Request{
		ID:        atomic.AddUint64(&prevRequestID, 1),
		Method:    method,
		RawParams: params,
	}
	c.rq.add(request, rhf)
	c.jsonOutChan <- request
}

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

type rqPair struct {
	request         *Request
	respHandlerFunc ResponseHandlerFunc
	put             time.Time
}

// TODO
type requestQ struct {
	sync.RWMutex
	pairs map[interface{}]*rqPair
}

func (rq *requestQ) watch() {

}

func (rq *requestQ) stopWatching() {

}

func (rq *requestQ) add(r *Request, rhf ResponseHandlerFunc) {
	rq.Lock()
	defer rq.Unlock()
	rq.pairs[r.ID] = &rqPair{
		request:         r,
		respHandlerFunc: rhf,
		put:             time.Now(),
	}
}

func (rq *requestQ) remove(id interface{}) (*rqPair, bool) {
	rq.Lock()
	defer rq.Unlock()
	pair, ok := rq.pairs[id]
	if ok {
		delete(rq.pairs, id)
	}
	return pair, ok
}

type RequestHandler interface {
	Handle(request *Request, rt ResponseTransmitter)
}

// ChannelConnected is published when websocket connection is established
// and channel is ready for interaction
type ChannelConnected struct {
	Time      time.Time `json:"time"`
	ChannelID string    `json:"channel"`
	Text      string    `json:"text"`
}

// ResponseTransmitter allows to respond to request from a certain handler.
// The implementation must guarantee that reply will be eventually made.
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

	// Channel returns the channel this transmitter belong to.
	Channel() Channel
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
