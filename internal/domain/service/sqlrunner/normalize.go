package sqlrunner

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"strconv"
	"time"
)

func normalizeDBValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(t)
	case time.Time:
		return t.Format(time.RFC3339Nano)
	case pgtype.Numeric:
		jsVal, _ := t.Float64Value()
		return strconv.FormatFloat(jsVal.Float64, 'f', -1, 64)
	default:
		if marshaler, ok := v.(json.Marshaler); ok {
			if jsonVal, err := marshaler.MarshalJSON(); err == nil {
				return string(jsonVal)
			}
		}
		return fmt.Sprint(v)
	}
}
