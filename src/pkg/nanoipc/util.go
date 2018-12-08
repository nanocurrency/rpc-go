package nanoipc

import "fmt"

// Error encapsulates a code, message and category.
// The Code is non-zero to indicate errors. Message and
// Category are usually set for errors transmitted by the node.
// This type implements the Go error interface.
type Error struct {
	Code     int
	Message  string
	Category string
}

// Returns an error string in the format ERRORCODE:CATEGORY:MESSAGE where
// the CATEGORY is optional.
func (e *Error) Error() string {
	if len(e.Category) > 0 {
		return fmt.Sprintf("%d:%s:%s", e.Code, e.Category, e.Message)
	} else {
		return fmt.Sprintf("%d:%s", e.Code, e.Message)
	}
}

// A CallChain allows safe chaining of functions. If an error
// occurs, the remaining functions in the chain are not called.
type CallChain struct {
	Err *Error
}

// Invokes the callback if there are no errors.
// The callback should set the CallChain#Err property on error */
func (sc *CallChain) Do(fn func()) *CallChain {
	if sc.Err == nil {
		fn()
	}
	return sc
}

// Invokes the callback if there's an error
func (sc *CallChain) Failure(fn func()) *CallChain {
	if sc.Err != nil {
		fn()
	}
	return sc
}
