package failure

type baseError struct {
	err error
}

func newBaseError(err error) baseError {
	return baseError{
		err: err,
	}
}

func (e baseError) Error() string {
	return e.err.Error()
}

func (e baseError) Unwrap() error {
	return e.err
}
