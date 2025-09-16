package errors

import "errors"

var (
	ErrBlueprintNotFound    = errors.New("blueprint file not found")
	ErrBlueprintParseFailed = errors.New("blueprint parsing failed")
	ErrScaffoldFailed       = errors.New("scaffolding failed")
	ErrSCMFailed           = errors.New("SCM operation failed")
	ErrProvisionFailed     = errors.New("provisioning failed")
	ErrRuntimeFailed       = errors.New("runtime operation failed")
	ErrConfigInvalid       = errors.New("configuration invalid")
	ErrNetworkFailed       = errors.New("network operation failed")
	ErrFileSystemFailed    = errors.New("filesystem operation failed")
)

type KloneKitError struct {
	Type        error
	Context     string
	Cause       string
	Suggestion  string
	OriginalErr error
}

func (e *KloneKitError) Error() string {
	return e.OriginalErr.Error()
}

func (e *KloneKitError) Unwrap() error {
	return e.OriginalErr
}

func NewKloneKitError(errorType error, context, cause, suggestion string, originalErr error) *KloneKitError {
	return &KloneKitError{
		Type:        errorType,
		Context:     context,
		Cause:       cause,
		Suggestion:  suggestion,
		OriginalErr: originalErr,
	}
}

func NewBlueprintError(context, cause, suggestion string, originalErr error) *KloneKitError {
	return NewKloneKitError(ErrBlueprintNotFound, context, cause, suggestion, originalErr)
}

func NewParseError(context, cause, suggestion string, originalErr error) *KloneKitError {
	return NewKloneKitError(ErrBlueprintParseFailed, context, cause, suggestion, originalErr)
}

func NewScaffoldError(context, cause, suggestion string, originalErr error) *KloneKitError {
	return NewKloneKitError(ErrScaffoldFailed, context, cause, suggestion, originalErr)
}

func NewSCMError(context, cause, suggestion string, originalErr error) *KloneKitError {
	return NewKloneKitError(ErrSCMFailed, context, cause, suggestion, originalErr)
}

func NewProvisionError(context, cause, suggestion string, originalErr error) *KloneKitError {
	return NewKloneKitError(ErrProvisionFailed, context, cause, suggestion, originalErr)
}

func NewRuntimeError(context, cause, suggestion string, originalErr error) *KloneKitError {
	return NewKloneKitError(ErrRuntimeFailed, context, cause, suggestion, originalErr)
}

func NewConfigError(context, cause, suggestion string, originalErr error) *KloneKitError {
	return NewKloneKitError(ErrConfigInvalid, context, cause, suggestion, originalErr)
}

func NewNetworkError(context, cause, suggestion string, originalErr error) *KloneKitError {
	return NewKloneKitError(ErrNetworkFailed, context, cause, suggestion, originalErr)
}

func NewFileSystemError(context, cause, suggestion string, originalErr error) *KloneKitError {
	return NewKloneKitError(ErrFileSystemFailed, context, cause, suggestion, originalErr)
}