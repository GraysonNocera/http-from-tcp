package response

import (
	"fmt"
	"http-from-tcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

type Writer struct {
  Writer io.Writer
  WriterState WriterState
}

type WriterState int
const (
  WriterStateStatusLine = iota
  WriterStateHeaders
  WriterStateBody
  WriterStateDone
)

const (
  StatusCode200 = 200
  StatusCode400 = 400
  StatusCode500 = 500
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
  if w.WriterState != WriterStateStatusLine {
    return fmt.Errorf("invalid, not in writer state")
  }
  err := WriteStatusLine(w.Writer, statusCode)
  if err != nil {
    return fmt.Errorf("error in writing status line: %w", err)
  }
  w.WriterState = WriterStateHeaders
  return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
  if w.WriterState != WriterStateHeaders {
    return fmt.Errorf("invalid state, not in header state")
  }
  err := WriteHeaders(w.Writer, headers)
  if err != nil {
    return fmt.Errorf("error in writing headers: %w", err)
  }
  w.WriterState = WriterStateBody
  return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
  if w.WriterState != WriterStateBody {
    return 0, fmt.Errorf("invalid state, not in body state")
  }
  n, err := w.Writer.Write(p)
  if err != nil {
    return 0, fmt.Errorf("error when writing body: %w", err)
  }
  w.WriterState = WriterStateDone
  return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
  totalWritten := 0
  for len(p) > 0 {
    var bytesInChunk byte = min(0xff, byte(len(p)))
    n, err := w.Writer.Write([]byte{bytesInChunk, headers.CRLF[0], headers.CRLF[1]})
    if err != nil {
      return 0, err
    }
    n, err = w.Writer.Write(p[:bytesInChunk])
    if err != nil {
      return 0, err
    }
    p = p[n:]

    n, err = w.Writer.Write([]byte(headers.CRLF))
    if err != nil {
      return 0, err
    }

    totalWritten += n
  }
  return totalWritten, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
  _, err := w.Writer.Write([]byte{0x00, headers.CRLF[0], headers.CRLF[1]})
  if err != nil {
    return 0, err
  }
  _, err = w.Writer.Write([]byte(headers.CRLF))
  if err != nil {
    return 0, err
  }
  return 0, nil
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
  reasonPhrase := ""
  version := "HTTP/1.1"
  switch statusCode {
  case StatusCode200:
    reasonPhrase = version + " " + strconv.Itoa(StatusCode200) + " " + "OK"
  case StatusCode400:
    reasonPhrase = version + " " + strconv.Itoa(StatusCode400) + " " + "Bad Request"
  case StatusCode500:
    reasonPhrase = version + " " + strconv.Itoa(StatusCode500) + " " + "Internal Server Error"
  default:
    reasonPhrase = version + " " + strconv.Itoa(int(statusCode)) + " "
  }
  reasonPhrase += "\r\n"
  // fmt.Printf("%s", reasonPhrase)
  _, err := w.Write([]byte(reasonPhrase));
  return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
  h := headers.NewHeaders()
  h.Set("Content-Length", strconv.Itoa(contentLen))
  h.Set("Connection", "close")
  h.Set("Content-Type", "text/html")
  return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
  h := ""
  for key, value := range headers {
    str := fmt.Sprintf("%s: %s\r\n", key, value)
    h += str
  }
  h += "\r\n"
  // fmt.Printf("%s", h)
  _, err := w.Write([]byte(h))
  return err
}
