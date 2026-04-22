package sqlrunner

import (
	"fmt"
	"time"
)

func NormalizeDBValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(t)
	case time.Time:
		return t.Format(time.RFC3339Nano)
	default:
		return fmt.Sprint(v)
	}
}
