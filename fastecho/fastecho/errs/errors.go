package errs

import (
	"net/http"
	"runtime"

	"github.com/pkg/errors"
)

// Inspired by: https://dangillis.dev/posts/errors/

// ErrorType defines the type of the error.
type ErrorType uint8

// Types of errors.
//
// The values of the error types are common for the whole app.
// Do not reorder this list or remove any items since that will
// change their values. New items must be added only to the end
const (
	Other        ErrorType = iota // Unclassified error
	NotFound                      // Item does not exist
	BadRequest                    // A remote REST call returned HTTP 400
	Unauthorized                  // Request is unauthorized to perform this call
)

// Error represents an error that has a type.
type Error struct {
	Type ErrorType
	Err  error
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.Err.Error()
}

// New builds an error value from its arguments. The type of each argument
// determines its meaning. If more than one argument of a given type is presented,
// only the last one is recorded.
//
// The types are:
//
//	string
//		Treated as an error message and assigned to the
//		Err field after a call to errors.New
//	errs.ErrorType
//		The class of error, such as not found
//	error
//		The underlying error that triggered this one
//
// If the error is printed, only those items that have
// been set to non-zero values will appear in the result
//
// If Type is not specified or Other, we set it to the Type of the underlying error
func New(args ...interface{}) error {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	if len(args) == 0 {
		return errors.New("an error occurred")
	}

	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			e.Err = errors.New(arg)
		case ErrorType:
			e.Type = arg
		case *Error:
			e.Err = arg
		case error:
			// if the error implements stackTracer, then it is
			// a pkg/errors error type and does not need to have
			// the stack added
			_, ok := arg.(stackTracer)
			if ok {
				e.Err = arg
			} else {
				e.Err = errors.WithStack(arg)
			}
		default:
			_, file, line, _ := runtime.Caller(1)
			return errors.Errorf("errors.New: bad call from %s:%d: %v, unknown type %T, value %v in error call", file, line, args, arg, arg)
		}
	}

	prev, ok := e.Err.(*Error)
	if !ok {
		return e
	}

	// If this error has Kind unset or Other, pull up the inner one.
	if e.Type == Other {
		e.Type = prev.Type
		prev.Type = Other
	}

	return e
}

// TypeIs reports whether err is an *Error of the given Type. If err is nil then TypeIs returns false.
func TypeIs(t ErrorType, err error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}

	if e.Type != Other {
		return e.Type == t
	}

	if e.Err != nil {
		return TypeIs(t, e.Err)
	}

	return false
}

// GetHTTPCode returns an HTTP status code that corresponds to the ErrorType.
func GetHTTPCode(err error) int {
	e, ok := err.(*Error)
	if !ok {
		return http.StatusInternalServerError
	}

	switch e.Type {
	case NotFound:
		return http.StatusNotFound
	case BadRequest:
		return http.StatusBadRequest
	case Unauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
