package errcodes

type Code string

func (e Code) String() string {
	return string(e)
}

const (
	Internal      Code = "InternalError"
	NotFound      Code = "NotFound"
	Validation    Code = "ValidationError"
	Unauthorized  Code = "Unauthorized"
	AlreadyExists Code = "AlreadyExists"
)
