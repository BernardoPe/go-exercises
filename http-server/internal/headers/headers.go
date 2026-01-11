package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(key string) (string, bool) {
	value, exists := h[strings.ToLower(key)]
	return value, exists
}

func (h Headers) Set(key, value string) {
	h[strings.ToLower(key)] = value
}

const (
	crlf = "\r\n"
)

var errInvalidHeader = errors.New("invalid header")

func parseFieldLine(line []byte) (key, value string, err error) {
	if len(line) == 0 {
		return "", "", errInvalidHeader
	}

	colon := bytes.IndexByte(line, ':')
	if colon <= 0 {
		return "", "", errInvalidHeader
	}

	fieldName := bytes.ToLower(bytes.TrimLeft(line[:colon], " \t"))

	if !isValidToken(fieldName) {
		return "", "", errInvalidHeader
	}

	fieldValue := bytes.TrimSpace(line[colon+1:])

	return string(fieldName), string(fieldValue), nil
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	processed := 0
	done = false

	for {
		idx := bytes.Index(data[processed:], []byte(crlf))
		if idx == -1 {
			break
		}

		line := data[processed : processed+idx]

		if len(line) == 0 {
			processed += len(crlf)
			done = true
			break
		}

		key, value, err := parseFieldLine(line)
		if err != nil {
			return 0, false, err
		}

		if _, exists := h[key]; exists {
			h[key] = h[key] + ", " + value
		} else {
			h[key] = value
		}

		processed += len(line) + len(crlf)
	}

	return processed, done, nil
}

func isValidToken(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	for _, c := range b {
		if !isTchar(c) {
			return false
		}
	}
	return true
}

func isTchar(c byte) bool {
	switch {
	case c >= '0' && c <= '9':
		return true
	case c >= 'A' && c <= 'Z':
		return true
	case c >= 'a' && c <= 'z':
		return true
	case strings.ContainsRune("!#$%&'*+-.^_`|~", rune(c)):
		return true
	default:
		return false
	}
}
