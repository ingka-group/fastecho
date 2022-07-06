package gcstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

type ClientMock struct {
	T        *testing.T
	Error    bool
	Response interface{}
}

func (c ClientMock) read(_ context.Context, _, _ string) ([]byte, error) {
	if c.Error {
		return nil, fmt.Errorf("error")
	}

	bb, err := json.Marshal(c.Response)
	if err != nil {
		c.T.Fail()
	}

	return bb, nil
}

func (c ClientMock) write(_ context.Context, _, _ string, _ []byte) error {
	if c.Error {
		return fmt.Errorf("error")
	}

	return nil
}

func (c ClientMock) delete(_ context.Context, _, _ string) error {
	if c.Error {
		return StorageObjectNotFound
	}

	return nil
}
