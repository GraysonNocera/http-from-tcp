package request

import (
	"bytes"
	"fmt"
	"http-from-tcp/internal/headers"
	"io"
	"strconv"
	"unicode"
)

type ParseState int

const (
	ParseStateRequestLine ParseState = iota
	ParseStateHeaders
  ParseStateBody
	ParseStateDone
	ParseStateError
)

type Request struct {
	RequestLine RequestLine
	Headers headers.Headers
	Body []byte
	state ParseState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const CRLF = "\r\n"
const BUFFER_SIZE = 8
const MAX_BUF_SIZE = 1024

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := Request{state: ParseStateRequestLine}
	r.Headers = headers.NewHeaders()
	buf := make([]byte, BUFFER_SIZE)
  readPos := 0
  writePos := 0
	for r.state != ParseStateDone {
    if writePos == len(buf) { // we have reached end of buffer
      if readPos > 0 { // we should discard the data before readPos, as it has been both read and processed
        copy(buf, buf[readPos:writePos])
        writePos -= readPos
        readPos = 0
      } else { // readPos at 0, meaning the entire buffer is full of unparsed data, we must grow it
        newBuf := make([]byte, len(buf) * 2)
        copy(newBuf, buf)
        buf = newBuf
      }
    }

		numRead, err := reader.Read(buf[writePos:]) // extend the middle portion (read but not processed)
		if err == io.EOF {
			r.state = ParseStateDone
			break;
		}
    writePos += numRead
		if err != nil {
			return nil, fmt.Errorf("error reading from reader: %w", err)
		}
		numParsed, err := r.parse(buf[readPos:writePos])
		if err != nil {
			return nil, fmt.Errorf("error parsing buffer: %w", err)
		}
    // fmt.Printf("total Parsed: %d\n", numParsed)
    // fmt.Printf("state: %d\n\n", r.state)
    readPos += numParsed
	}
	return &r, nil
}

func (r *Request) parse(data []byte) (int, error) {
  totalParsed := 0
  for {
    // fmt.Printf("data: '%s'\n", string(data));
    switch r.state {
    case ParseStateRequestLine:
      // fmt.Printf("data: '%s'\n", string(data));
      requestLine, n, err := parseRequestLine(data)
      if err != nil {
        return 0, fmt.Errorf("error parsing request line: %w", err)
      }
      if n == 0 { // need more data
        return totalParsed, nil
      }
      r.state = ParseStateHeaders
      r.RequestLine = *requestLine
      data = data[n:]
      totalParsed += n
      // fmt.Printf("parsed line: '%s'\n", r.RequestLine)
    case ParseStateHeaders:
      // fmt.Printf("data: '%s'\n", string(data));
      n, done, err := r.Headers.Parse(data)
      if err != nil {
        return 0, fmt.Errorf("error parsing headers, %w", err)
      }
      if n == 0 { // need more data
        return totalParsed, nil
      }
      totalParsed += n
      data = data[n:]
      if done {
        r.state = ParseStateBody
      }
    case ParseStateBody:
      // fmt.Printf("data: '%s'\n", string(data));
      contentLength := r.Headers.Get("Content-Length")
      // fmt.Printf("content length: %s\n", contentLength)
      if contentLength == "" {
        r.state = ParseStateDone
        return totalParsed, nil
      }
      length, err := strconv.Atoi(contentLength)
      if err != nil {
        return 0, fmt.Errorf("error parsing content-length: %w", err)
      }
      remaining := min(length - len(r.Body), len(data))
      r.Body = append(r.Body, data[:remaining]...)
      if len(r.Body) > length {
        return 0, fmt.Errorf("body does not match content-length: %d > %d, '%s'", len(data), length, string(data))
      }
      if len(r.Body) == length {
        r.state = ParseStateDone
      }
      totalParsed += remaining
      return totalParsed, nil
    }
  }
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
  // fmt.Printf("data: '%s'\n", string(data))

  ind := bytes.Index(data, []byte(CRLF))
	if ind == -1 {
		return nil, 0, nil // need more data, haven't found CRLF
	}

	line := data[:ind]
  // fmt.Printf("line: '%s'\n", string(line))
  parts := bytes.Split(line, []byte(" "))
	if (len(parts) != 3) {
		return nil, 0, fmt.Errorf("malformed request line: %s", line)
	}

	method := parts[0]
  // fmt.Printf("method: %s\n", string(method))
	if !isAllCapitalAlphabetCharacters(method) {
		return nil, 0, fmt.Errorf("malformed method: %s", method)
	}

	requestTarget := parts[1]
  // fmt.Printf("requestTarget: %s\n", string(requestTarget))
	if !bytes.HasPrefix(requestTarget, []byte("/")) {
		return nil, 0, fmt.Errorf("malformed request target: %s", requestTarget)
	}

	version, err := parseRequestLineVersion(parts[2])
	if err != nil {
		return nil, 0, fmt.Errorf("malformed version: %s, %w", version, err)
	}

	requestLine := RequestLine{HttpVersion: string(version), RequestTarget: string(requestTarget), Method: string(method)}

	return &requestLine, len(line) + len(CRLF), nil
}

func parseRequestLineVersion(s []byte) ([]byte, error) {
  // fmt.Printf("line version: %s\n", s)
	if !bytes.HasPrefix(s, []byte("HTTP/")) {
		return nil, fmt.Errorf("version does not contain http prefix")
	}
	parts := bytes.Split(s, []byte("/"))
	if len(parts) != 2 {
		return nil, fmt.Errorf("version has too many parts split by '/'")
	}
  // fmt.Printf("parts[0]: '%s'\n", parts[0])
  // fmt.Printf("parts[1]: '%s'\n", parts[1])
	if !bytes.Equal(parts[1], []byte("1.1")) {
		return nil, fmt.Errorf("HTTP version is not 1.1")
	}
	version := bytes.Split(s, []byte("HTTP/"))[1]
	return version, nil
}

func isAllCapitalAlphabetCharacters(s []byte) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(rune(r)) {
			return false
		}
		if !unicode.IsUpper(rune(r)) {
			return false
		}
	}
	return true
}