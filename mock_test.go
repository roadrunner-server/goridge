package goridge

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	read int = iota
	write
)

type dataSegment struct {
	pipe int
	data []byte
}

type connMock struct {
	closed   bool
	position int
	expects  []*dataSegment
}

func (m *connMock) expect(pipe int, data []byte) {
	m.expects = append(m.expects, &dataSegment{
		pipe: pipe,
		data: data,
	})
}

func (m *connMock) Read(p []byte) (n int, err error) {
	next, err := m.nextSegment(read)
	if err != nil {
		return 0, err
	}

	copy(p, next)

	return len(next), nil
}

func (m *connMock) Write(p []byte) (n int, err error) {
	next, err := m.nextSegment(write)
	if err != nil {
		return 0, err
	}

	if !bytes.Equal(next, p) {
		return 0, errors.New("payload expectation mismatch")
	}

	return len(next), nil
}

func (m *connMock) Close() error {
	m.closed = true
	return nil
}

func (m *connMock) nextSegment(pipe int) ([]byte, error) {
	if len(m.expects) <= m.position {
		return nil, fmt.Errorf("unable to find data segment on position %v", m.position)
	}

	segment := m.expects[m.position]
	if segment.pipe != pipe {
		return nil, fmt.Errorf("pipe mismatch %v / %v", segment.pipe, pipe)
	}

	m.position++

	return segment.data, nil
}

func (m *connMock) leftSegments() int {
	return len(m.expects) - m.position
}
