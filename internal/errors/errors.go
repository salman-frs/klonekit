package errors

import "sync"

var (
	defaultHandler *ErrorHandler
	once           sync.Once
)

func GetDefaultHandler() (*ErrorHandler, error) {
	var err error
	once.Do(func() {
		defaultHandler, err = NewErrorHandler()
	})
	return defaultHandler, err
}

func HandleError(err error) {
	if handler, handlerErr := GetDefaultHandler(); handlerErr == nil {
		handler.Handle(err)
	}
}

// resetDefaultHandler resets the singleton for testing purposes
func resetDefaultHandler() {
	defaultHandler = nil
	once = sync.Once{}
}