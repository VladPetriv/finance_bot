package errs

// Err represents a custom error type with a message.
type Err struct { //nolint:errname
	Message string `json:"message"`
}

var _ error = (*Err)(nil)

// New creates a new custom error with the given message.
func New(message string) *Err {
	return &Err{Message: message}
}

func (e *Err) Error() string {
	return e.Message
}

// IsExpected checks if the given error is of custom Err type.
func IsExpected(err error) bool {
	_, ok := err.(*Err) //nolint:errorlint
	return ok
}
