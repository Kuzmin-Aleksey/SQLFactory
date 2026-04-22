package middlwarex

import (
	"math"
	"unicode"
)

func jsonUnmarshal(p []byte) any {
	v, _ := unmarshalValue(p)
	return v
}

func unmarshalValue(p []byte) (any, int) {
	var v any

	for i, r := range p {
		switch r {
		case '{':
			v, n := unmarshalStruct(p[i:])
			return v, n + i

		case '}':
			return v, i

		case '[':
			v, n := unmarshalSlice(p[i:])
			return v, n + i

		case '"', '\'':
			v, n := unmarshalStringVal(p[i:])
			return v, n + i

		case ' ':

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			v, n := unmarshalDigitVal(p[i:])
			return v, n + i

		case 't':
			return true, min(i+3, len(p))
		case 'f':
			return false, min(i+4, len(p))
		case 'n':
			return nil, min(i+3, len(p))

		}
	}

	return v, len(p)
}

func unmarshalStringVal(p []byte) (string, int) {
	if len(p) < 1 {
		return "", len(p)
	}
	q := p[0]

	for i := 1; i < len(p); i++ {
		if p[i] == q && p[i-1] != '\\' {
			return string(p[1:i]), i
		}
	}
	return string(p[1:]) + "EOF", len(p)
}

func unmarshalDigitVal(p []byte) (any, int) {
	var v int

	isNegative := 0
	isParsingFloat := false
	floatPart := 0.0

	if p[0] == '-' {
		isNegative = 1
	}
	var (
		i  int
		i1 int
	)

	for i = isNegative; i < len(p); i++ {
		r := p[i]
		if r == '.' {
			isParsingFloat = true
			continue
		}
		if !unicode.IsDigit(rune(r)) {
			break
		}

		if isParsingFloat {
			i1++
			floatPart += float64(r-'0') / math.Pow(10, float64(i1))
		} else {
			v = v*10 + int(r-'0')
		}

	}

	if isParsingFloat {
		return (float64(v) + floatPart) * float64(1-2*(isNegative%2)), i
	}

	return v * (1 - 2*(isNegative%2)), i
}

func unmarshalStruct(p []byte) (map[string]any, int) {
	findingKey := true
	keyIdx1 := 0
	scanKey := false

	m := make(map[string]any)

	for i := 0; i < len(p); i++ {
		r := p[i]

		if r == '}' {
			return m, i
		}

		if findingKey {
			if r == '"' || r == '\'' {
				findingKey = false
				scanKey = true
				keyIdx1 = i + 1
			}
			continue
		}
		if scanKey {
			if r == '"' || r == '\'' {
				scanKey = false
				findingKey = true

				if i == len(p)-1 {
					m[string(p[keyIdx1:i])] = "EOF"
					return m, len(p)
				}

				v, n := unmarshalValue(p[i+1:])
				m[string(p[keyIdx1:i])] = v
				i += n + 1
			}
		}
	}

	return m, len(p)
}

func unmarshalSlice(p []byte) ([]any, int) {
	s := make([]any, 0)

	if len(p) < 1 {
		return s, 1
	}

	findingComma := false

	for i := 1; i < len(p); i++ {
		r := p[i]

		if r == ']' {
			return s, i + 1
		}
		if findingComma {
			if r == ',' {
				findingComma = false
			}
			continue
		}

		v, n := unmarshalValue(p[i:])
		s = append(s, v)
		i += n - 1

		findingComma = true
	}

	return s, len(p)
}
