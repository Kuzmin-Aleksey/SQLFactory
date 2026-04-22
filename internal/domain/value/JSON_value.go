package value

type JsonValue string

func (v *JsonValue) String() string {
	return string(*v)
}

func (v *JsonValue) MarshalJSON() ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	return []byte(*v), nil
}

func (v *JsonValue) UnmarshalJSON(data []byte) error {
	*v = JsonValue(data)
	return nil
}
