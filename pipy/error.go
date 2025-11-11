package pipy

import "fmt"

type Error struct {
	Message string
	Code    int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %d", e.Message, e.Code)
}

var (
	RepoNotFound = &Error{Message: "not found", Code: 404}
)

func newError(format string, a ...any) *Error {
	errMessage := fmt.Sprintf(format, a...)
	Logger.Error(errMessage)
	return &Error{Message: errMessage, Code: 500}
}
