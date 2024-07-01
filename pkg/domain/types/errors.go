package types

import "net/http"

type Error struct {
	code  int
	msg   string
	cause error
}

func (x Error) Error() string {
	msg := x.msg
	if x.cause != nil {
		msg += ": " + x.cause.Error()
	}
	return msg
}
func (x Error) Code() int { return x.code }
func (x Error) Wrap(cause error) Error {
	return Error{code: x.code, msg: x.msg, cause: cause}
}
func (x Error) Unwrap() error { return x.cause }

var (
	ErrInvalidContentType = Error{code: http.StatusBadRequest, msg: "unsupported Content-Type"}
	ErrInvalidInput       = Error{code: http.StatusBadRequest, msg: "invalid input"}
	ErrAuthFailed         = Error{code: http.StatusUnauthorized, msg: "authentication failed"}
	ErrForbidden          = Error{code: http.StatusForbidden, msg: "forbidden"}
)
