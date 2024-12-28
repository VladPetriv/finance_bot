package errs

type Err struct { //nolint:errname
	Message string `json:"message"`
}

var _ error = (*Err)(nil)

func New(message string) *Err {
	return &Err{Message: message}
}

func (e *Err) Error() string {
	return e.Message
}

func IsExpected(err error) bool {
	_, ok := err.(*Err) //nolint:errorlint
	return ok
}
