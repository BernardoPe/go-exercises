package response

import (
	"errors"
	"fmt"
	"http_server/internal/headers"
	"io"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	Created             StatusCode = 201
	NoContent           StatusCode = 204
	BadRequest          StatusCode = 400
	NotFound            StatusCode = 404
	MethodNotAllowed    StatusCode = 405
	InternalServerError StatusCode = 500
)

type writerState int

const (
	stateStatusLine writerState = iota
	stateHeaders
	stateBody
	stateTrailers
	stateDone
)

type Writer struct {
	conn  io.Writer
	state writerState
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{
		conn:  conn,
		state: stateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != stateStatusLine {
		return errors.New("WriteStatusLine must be called first")
	}

	var statusText string
	switch statusCode {
	case OK:
		statusText = "OK"
	case Created:
		statusText = "Created"
	case NoContent:
		statusText = "No Content"
	case BadRequest:
		statusText = "Bad Request"
	case NotFound:
		statusText = "Not Found"
	case MethodNotAllowed:
		statusText = "Method Not Allowed"
	case InternalServerError:
		statusText = "Internal Server Error"
	default:
		statusText = "Unknown Status"
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, statusText)
	_, err := w.conn.Write([]byte(statusLine))
	if err != nil {
		return err
	}
	w.state = stateHeaders
	return nil
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.state != stateHeaders {
		return errors.New("WriteHeaders must be called after WriteStatusLine")
	}
	for k, v := range h {
		headerLine := fmt.Sprintf("%s: %s\r\n", k, v)
		_, err := w.conn.Write([]byte(headerLine))
		if err != nil {
			return err
		}
	}
	_, err := w.conn.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.state = stateBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateBody {
		return 0, errors.New("WriteBody must be called after WriteHeaders")
	}

	n, err := w.conn.Write(p)
	if err != nil {
		return n, err
	}
	w.state = stateDone
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != stateBody {
		return 0, errors.New("WriteChunkedBody must be called after WriteHeaders")
	}
	if len(p) == 0 {
		return 0, nil
	}
	chunkSize := fmt.Sprintf("%X\r\n", len(p))
	_, err := w.conn.Write([]byte(chunkSize))
	if err != nil {
		return 0, err
	}

	n, err := w.conn.Write(p)
	if err != nil {
		return n, err
	}

	_, err = w.conn.Write([]byte("\r\n"))
	if err != nil {
		return n, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != stateBody {
		return 0, errors.New("WriteChunkedBodyDone must be called after WriteHeaders")
	}

	n, err := w.conn.Write([]byte("0\r\n"))
	if err != nil {
		return n, err
	}
	w.state = stateTrailers
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != stateTrailers {
		return errors.New("WriteTrailers must be called after WriteChunkedBodyDone")
	}
	for k, v := range h {
		trailerLine := fmt.Sprintf("%s: %s\r\n", k, v)
		_, err := w.conn.Write([]byte(trailerLine))
		if err != nil {
			return err
		}
	}
	_, err := w.conn.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.state = stateDone
	return nil
}

func GetDefaultHeaders(contentLength int) headers.Headers {
	h := headers.NewHeaders()
	h["content-type"] = "text/plain"
	h["content-length"] = fmt.Sprintf("%d", contentLength)
	h["connection"] = "close"
	return h
}
