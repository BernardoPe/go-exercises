package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:80\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:80", headers["host"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("   Host:    localhost:80    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:80", headers["host"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	headers["user-agent"] = "curl/7.81.0"
	data = []byte("Host: localhost:80\r\nAccept: */*\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:80", headers["host"])
	assert.Equal(t, "*/*", headers["accept"])
	assert.Equal(t, "curl/7.81.0", headers["user-agent"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Valid multiple values for same header
	headers = NewHeaders()
	data = []byte("Set-Cookie: id=123\r\nSet-Cookie: token=abc\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "id=123, token=abc", headers["set-cookie"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Incomplete headers
	headers = NewHeaders()
	data = []byte("Host: localhost:80\r\nAccept: */*")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("Host: localhost:80\r\nAccept: */*\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Invalid header format (missing colon)
	headers = NewHeaders()
	data = []byte("Host localhost80\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Empty header key
	headers = NewHeaders()
	data = []byte(": localhost:80\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:80       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid header key with whitespace
	headers = NewHeaders()
	data = []byte("Test Key: value\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
