package request

import (
	"bytes"
	"fmt"
	headers "http_server/internal/headers"
	"io"
	"strings"
)

type Request struct {
	Line    Line
	Headers headers.Headers
	Body    Body
	state   State
}

type Line struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Body struct {
	Data   []byte
	Length int
}

type State int

const (
	ParseLine = iota
	ParseHeaders
	ParseBody
	Done
)

func FromReader(reader io.Reader) (*Request, error) {
	req := NewRequest()
	buf := make([]byte, 1024)
	bufLen := 0

	for !req.Done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil && err != io.EOF {
			return nil, err
		}
		bufLen += n
		if err == io.EOF {
			break
		}

		if bufLen == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		processed, parseErr := req.parse(buf[:bufLen])
		if parseErr != nil {
			return nil, parseErr
		}

		if processed > 0 {
			copy(buf, buf[processed:bufLen])
			bufLen -= processed
		}
	}

	if !req.Done() {
		return nil, fmt.Errorf("incomplete request")
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	processed := 0
	for {
		switch r.state {
		case ParseLine:
			l, n, err := parseRequestLine(data)
			if err != nil {
				return processed, err
			}
			if n == 0 {
				return processed, nil
			}
			processed += n
			r.Line = *l
			r.state = ParseHeaders
		case ParseHeaders:
			n, done, err := r.Headers.Parse(data[processed:])
			if err != nil {
				return processed, err
			}
			processed += n
			if !done {
				return processed, nil
			}
			r.state = ParseBody
		case ParseBody:
			n, done, err := r.parseBody(data[processed:])
			if err != nil {
				return processed, err
			}
			processed += n
			if !done {
				return processed, nil
			}
			r.state = Done
			return processed, nil
		case Done:
			return processed, nil
		default:
			return processed, nil
		}
	}
}

func (r *Request) parseBody(data []byte) (toRead int, done bool, err error) {
	if r.Body.Length == 0 {
		v, exists := r.Headers.Get("content-length")
		if !exists {
			return 0, true, nil
		}

		var contentLength int
		if _, err := fmt.Sscanf(v, "%d", &contentLength); err != nil {
			return 0, false, fmt.Errorf("invalid Content-Length value")
		}

		r.Body.Length = contentLength
		r.Body.Data = make([]byte, 0, contentLength)
	}

	remaining := r.Body.Length - len(r.Body.Data)
	if remaining <= 0 {
		return 0, true, nil
	}

	toRead = remaining
	if available := len(data); available < toRead {
		toRead = available
	}

	r.Body.Data = append(r.Body.Data, data[:toRead]...)

	done = len(r.Body.Data) == r.Body.Length

	return toRead, done, nil
}

const (
	Separator             = "\r\n"
	ErrInvalidLineFormat  = "invalid request-line format"
	ErrUnsupportedMethod  = "unsupported HTTP method"
	ErrUnsupportedVersion = "unsupported HTTP version"
)

func parseRequestLine(request []byte) (*Line, int, error) {
	lineEnd := bytes.Index(request, []byte(Separator))
	if lineEnd == -1 {
		return nil, 0, nil
	}

	line := request[:lineEnd]

	parts := bytes.SplitN(line, []byte(" "), 3)
	if len(parts) != 3 {
		return nil, 0, fmt.Errorf(ErrInvalidLineFormat)
	}

	l := &Line{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   strings.TrimPrefix(string(parts[2]), "HTTP/"),
	}

	if !l.ValidMethod() {
		return nil, 0, fmt.Errorf(ErrUnsupportedMethod)
	}

	if !l.ValidHttpVersion() {
		return nil, 0, fmt.Errorf(ErrUnsupportedVersion)
	}

	return l, lineEnd + len(Separator), nil
}

func (l *Line) ValidMethod() bool {
	switch l.Method {
	case "GET":
		return true
	case "HEAD":
		return true
	case "POST":
		return true
	case "PUT":
		return true
	case "DELETE":
		return true
	case "CONNECT":
		return true
	case "OPTIONS":
		return true
	case "TRACE":
		return true
	case "PATCH":
		return true
	default:
		return false
	}
}

func (l *Line) ValidHttpVersion() bool {
	return l.HttpVersion == "1.1"
}

func (r *Request) Done() bool {
	return r.state == Done
}

func NewRequest() *Request {
	return &Request{
		Line:    Line{},
		Headers: headers.NewHeaders(),
		state:   ParseLine,
	}
}
