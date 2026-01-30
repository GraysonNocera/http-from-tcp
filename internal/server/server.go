package server

import (
	// "bytes"
	// "fmt"
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"io"
	"log"
	"net"
	"strconv"
)

type HandlerError struct {
  StatusCode response.StatusCode
  Message string
}

// type Handler func(w io.Writer, req *request.Request) *HandlerError
type Handler func(w *response.Writer, req *request.Request)

type Server struct {
  listener net.Listener
  closed bool
}

func Serve(port int, handler Handler) (*Server, error) {
  listener, err := net.Listen("tcp", string(":" + strconv.Itoa(port)))
  server := &Server{listener: listener, closed: false}
  if err != nil {
    return nil, err
  }
  go server.listen(handler)
  return server, nil
}

func (s *Server) Close() error {
  return s.listener.Close()
}

func (s *Server) listen(h Handler) {
  for {
    conn, err := s.listener.Accept()
    if err != nil {
      log.Fatal("error listening: %s", err.Error())
    }
    if s.closed {
      return
    }

    go s.handle(conn, h)
  }
}

func (s *Server) handle(conn net.Conn, h Handler) {
  defer conn.Close()

	r, err := request.RequestFromReader(conn)
  if err != nil {
    log.Fatal("error getting request: ", err)
  }
  w := response.Writer{
    Writer: conn,
    WriterState: response.WriterStateStatusLine,
  }
  h(&w, r)
  // err = response.WriteStatusLine(conn, response.StatusCode200)
  // if err != nil {
  //   log.Fatal("error writing status line")
  // }
  // headers := response.GetDefaultHeaders(b.Len())
  // err = response.WriteHeaders(conn, headers)
  // if err != nil {
  //   log.Fatal("error writing headers", err)
  // }
  // // fmt.Printf("%s", b.Bytes())
  // _, err = conn.Write(b.Bytes())
  // if err != nil {
  //   log.Fatal("error writing response body", err)
  // }
}

func writeHandlerError(w io.Writer, h *HandlerError) error {
  err := response.WriteStatusLine(w, h.StatusCode)
  if err != nil {
    log.Fatal("error writing status line", err)
  }
  headers := response.GetDefaultHeaders(len(h.Message))
  err = response.WriteHeaders(w, headers)
  if err != nil {
    log.Fatal("error writing headers", err)
  }
  // fmt.Printf("%s\n", h.Message)
  _, err = w.Write([]byte(h.Message))
  if err != nil {
    log.Fatal("error writing headers", err)
  }
  return err
}