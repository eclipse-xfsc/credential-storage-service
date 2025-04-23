package tests

import (
	"context"

	"github.com/eclipse-xfsc/credential-storage-service/internal/connection"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/mock"
)

type SessionMock struct {
	mock.Mock
}

func (s *SessionMock) Closed() bool {
	args := s.Called()
	return args.Bool(0)
}

func (s *SessionMock) Close() {
}

func (s *SessionMock) Query(stmt string, values ...interface{}) connection.QueryInterface {
	args := s.Called(stmt, values)
	return args.Get(0).(connection.QueryInterface)
}

type QueryMock struct {
	mock.Mock
}

func (q *QueryMock) Consistency(consistency gocql.Consistency) connection.QueryInterface {
	return q
}

func (q *QueryMock) Exec() error {
	return nil
}

// Scan wraps the query's Scan method
func (q *QueryMock) Scan(dest ...interface{}) error {
	return nil
}

func (q *QueryMock) WithContext(c context.Context) connection.QueryInterface {
	return q
}
