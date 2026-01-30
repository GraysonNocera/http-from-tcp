package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParsing(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("       Host:       localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, len(data) - 2, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	data = []byte("Host:localhost:42069\r\nMoreHeader:localhostagain\r\n\r\n")
	n, done, err = headers.Parse(data)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 22, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data[n:])
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "localhostagain", headers.Get("MoreHeader"))
	assert.Equal(t, 27, n)
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	assert.True(t, done)

	// Test: Invalid character in header key
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)

	// Test: Duplicate field-name
	headers = NewHeaders()
	data = []byte("Host:localhost:1\r\nHost:localhost:2\r\nHost:localhost:3\r\n\r\n")
	n, done, err = headers.Parse(data)
	data = data[n:]
	n, done, err = headers.Parse(data)
	data = data[n:]
	n, done, err = headers.Parse(data)
	assert.Equal(t, "localhost:1, localhost:2, localhost:3", headers.Get("Host"))
}