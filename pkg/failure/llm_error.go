package failure

import "errors"

type LLMError struct {
	LLMMessage string
}

func NewLLMError(msg string) error {
	return LLMError{
		LLMMessage: msg,
	}
}

func (err LLMError) Error() string {
	return "llm error: " + err.LLMMessage
}

func IsLLMError(err error) bool {
	return errors.As(err, new(LLMError))
}
