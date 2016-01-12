package vimages

type Response struct {
	Fail     bool
	Response Payload
	Delete   []string
	Send     string
}

type Payload interface {
}

type PayloadFail struct {
	Code int    `json:"error_code"`
	Msg  string `json:"error_message"`
}

var (
	respInternalError     = NewResponseFail(3000, "vimages internal error")
	respUnsupportedMethod = NewResponseFail(3001, "unsupported http method")
	respBadTarget         = NewResponseFail(3002, "no such target in database")
	respBadPut            = NewResponseFail(3003, "can't put on a folder without a valid command parameter")
	respCollission        = NewResponseFail(3004, "can't post here without overwrite order")

	respPrivilegesError    = NewResponseFail(2000, "privileges internal error")
	respInvalidCredentials = NewResponseFail(2001, "invalid login credentials")
	respAccessDenied       = NewResponseFail(2002, "insufficient privileges to perform action")
)

func NewResponseFail(code int, msg string) *Response {

	response := new(Response)
	payload := new(PayloadFail)
	response.Response = payload
	response.Fail = true
	response.Delete = make([]string, 0)
	payload.Code = code
	payload.Msg = msg

	return response

}
