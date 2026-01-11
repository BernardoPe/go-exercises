package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	r, err := FromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.Line.Method)
	assert.Equal(t, "/", r.Line.RequestTarget)
	assert.Equal(t, "1.1", r.Line.HttpVersion)

	// Test: Good GET Request line with path
	r, err = FromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.Line.Method)
	assert.Equal(t, "/coffee", r.Line.RequestTarget)
	assert.Equal(t, "1.1", r.Line.HttpVersion)

	// Test: Good POST Request line with path
	r, err = FromReader(strings.NewReader("POST /submit HTTP/1.1\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.Line.Method)
	assert.Equal(t, "/submit", r.Line.RequestTarget)
	assert.Equal(t, "1.1", r.Line.HttpVersion)

	// Test: Invalid HTTP version
	_, err = FromReader(strings.NewReader("GET / HTTP/2.0\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid method
	_, err = FromReader(strings.NewReader("FETCH / HTTP/1.1\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid number of parts in request line
	_, err = FromReader(strings.NewReader("GET /HTTP/1.1\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Out of order parts in request line
	_, err = FromReader(strings.NewReader("HTTP/1.1 GET /\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid number of parts in request line
	_, err = FromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.Line.Method)
	assert.Equal(t, "/", r.Line.RequestTarget)
	assert.Equal(t, "1.1", r.Line.HttpVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.Line.Method)
	assert.Equal(t, "/coffee", r.Line.RequestTarget)
	assert.Equal(t, "1.1", r.Line.HttpVersion)

	// Test: Good POST Request line with path
	reader = &chunkReader{
		data:            "POST /submit HTTP/1.1\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 5,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.Line.Method)
	assert.Equal(t, "/submit", r.Line.RequestTarget)
	assert.Equal(t, "1.1", r.Line.HttpVersion)
}

func TestHeadersParse(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:80", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:80\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = FromReader(reader)
	require.Error(t, err)

	// Test: Empty Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, 0, len(r.Headers))

	// Test: Duplicate Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nSet-Cookie: id=123\r\nSet-Cookie: token=abc\r\n\r\n",
		numBytesPerRead: 4,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "id=123, token=abc", r.Headers["set-cookie"])

	// Test: Missing End of Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:80\r\nUser-Agent: curl/7.81.0\r\nAccept: */*",
		numBytesPerRead: 2,
	}
	r, err = FromReader(reader)
	require.Error(t, err)
}

func TestBodyParse(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:80\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.NoError(t, err)
	assert.Equal(t, "hello world!\n", string(r.Body.Data))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:80\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = FromReader(reader)
	require.Error(t, err)

	// Test: Empty Body, Content-Length 0
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:80\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.NoError(t, err)
	assert.Equal(t, 0, len(r.Body.Data))

	// Test: Empty Body, No Content-Length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:80\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.NoError(t, err)
	assert.Equal(t, 0, len(r.Body.Data))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:80\r\n" +
			"Content-Length: 10\r\n" +
			"\r\n" +
			"short",
		numBytesPerRead: 2,
	}
	r, err = FromReader(reader)
	require.Error(t, err)

	// Test: Body longer than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:80\r\n" +
			"Content-Length: 5\r\n" +
			"\r\n" +
			"extralongbodycontent",
		numBytesPerRead: 4,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "extra", string(r.Body.Data))

	// Test: No Content-Length with Body, should result in empty body
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:80\r\n" +
			"\r\n" +
			"bodywithoutlength",
		numBytesPerRead: 5,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, 0, len(r.Body.Data))
}
