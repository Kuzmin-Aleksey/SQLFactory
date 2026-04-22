package failure

import "errors"

type ExternalDBError struct {
	baseError
}

func NewExternalDBError(err error) error {
	return ExternalDBError{
		baseError: newBaseError(err),
	}
}

func (err ExternalDBError) Error() string {
	return "external db error: " + err.baseError.Error()
}

func IsExternalDBError(err error) bool {
	return errors.As(err, new(ExternalDBError))
}
