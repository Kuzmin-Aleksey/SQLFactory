package sqlrunner

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockConn struct {
	Conn
}

func (m MockConn) Query(context.Context, string) (*QueryResult, error) {
	return &QueryResult{}, nil
}

func TestStmtFilter_Query(t *testing.T) {
	tests := []struct {
		sql   string
		check func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool
	}{
		{
			sql:   "SELECT * FROM users",
			check: assert.NoError,
		},
		{
			sql:   "INSERT INTO users VALUES (1, 'test')",
			check: assert.Error,
		},
	}

	conn := withStmtFilter(MockConn{})

	for _, test := range tests {
		_, err := conn.Query(context.Background(), test.sql)
		test.check(t, err)
	}
}
