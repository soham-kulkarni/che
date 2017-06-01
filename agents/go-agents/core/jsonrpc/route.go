package jsonrpc

var (
	DefaultRouter = &Router{}
)

// Route defines named operation and its handler.
type Route struct {
	// Method is the operation name like defined by Request.Method.
	Method string

	// DecoderFunc used for decoding raw request parameters
	// into the certain object. If decoding is okay, then
	// decoded value will be passed to the HandlerFunc
	// of this request route, so it is up to the route
	// - to define type safe couple of DecoderFunc & HandlerFunc.
	DecoderFunc func(body []byte) (interface{}, error)

	// HandlerFunc handler for decoded request parameters.
	// If handler function can't perform the operation then
	// handler function should either return an error, or
	// send it directly within transmitter#SendError func.
	// Params is a value returned from the DecoderFunc.
	HandlerFunc func(params interface{}, t ResponseTransmitter)
}

type Router struct {

}

func (r *Router) Manage(conn NativeConn) {

}

func (r *Router) Handle(request *Request, rt ResponseTransmitter) {

}
