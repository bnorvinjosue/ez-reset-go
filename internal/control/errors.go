// Package control: backend errors.
package control

// BackendError is returned for backend-level failures.
type BackendError struct {
	Msg string
}

func (e *BackendError) Error() string {
	return e.Msg
}

// NewBackendError creates a BackendError with the given message.
func NewBackendError(msg string) *BackendError {
	return &BackendError{Msg: msg}
}
