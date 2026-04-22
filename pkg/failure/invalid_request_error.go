package failure

import "errors"

type InvalidRequestError struct {
	baseError
}

func NewInvalidRequestError(err error) error {
	return InvalidRequestError{
		baseError: newBaseError(err),
	}
}

func (err InvalidRequestError) Error() string {
	return "invalid request error: " + err.baseError.Error()
}

func IsInvalidRequestError(err error) bool {
	return errors.As(err, new(InvalidRequestError))
}
