package main

import (
	"crypto/sha256"
	"fmt"
	"http-from-tcp/internal/headers"
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"http-from-tcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)



const port = 42069

func main() {
//   var h server.Handler = func(w *response.Writer, req *request.Request) {
//     if req.RequestLine.RequestTarget == "/yourproblem" {
//       body := `
// <html>
//   <head>
//     <title>400 Bad Request</title>
//   </head>
//   <body>
//     <h1>Bad Request</h1>
//     <p>Your request honestly kinda sucked.</p>
//   </body>
// </html>
//       `
//       w.WriteStatusLine(response.StatusCode400)
//       w.WriteHeaders(response.GetDefaultHeaders(len(body)))
//       w.WriteBody([]byte(body))
//     }
//     if req.RequestLine.RequestTarget == "/myproblem" {
//       w.WriteStatusLine(response.StatusCode500)
//       body := `
// <html>
//   <head>
//     <title>500 Internal Server Error</title>
//   </head>
//   <body>
//     <h1>Internal Server Error</h1>
//     <p>Okay, you know what? This one is on me.</p>
//   </body>
// </html>
//       `
//       w.WriteHeaders(response.GetDefaultHeaders(len(body)))
//       w.WriteBody([]byte(body))
//     }
//     body := `
// <html>
//   <head>
//     <title>200 OK</title>
//   </head>
//   <body>
//     <h1>Success!</h1>
//     <p>Your request was an absolute banger.</p>
//   </body>
// </html>
//     `
//     w.WriteStatusLine(response.StatusCode200)
//     w.WriteHeaders(response.GetDefaultHeaders(len(body)))
//     w.WriteBody([]byte(body))
//   }

  var newH server.Handler = func(w *response.Writer, req *request.Request) {
    path := req.RequestLine.RequestTarget
    h := headers.NewHeaders()
    h.Set("Connection", "close")
    h.Set("Transfer-Encoding", "chunked")
    h.Set("Content-Type", "text/plain")
    h.Set("Trailer", "X-Content-SHA256,X-Content-Length")

    fullData := []byte{}
    if strings.HasPrefix(path, "/httpbin/") {
      newPath := strings.TrimPrefix(path, "/httpbin/")
      url := fmt.Sprintf("https://httpbin.org/%s", newPath)
      r, err := http.Get(url)
      if err != nil {
        log.Fatal("error getting response from httbin.org", err)
      }
      w.WriteStatusLine(200)
      w.WriteHeaders(h)
      for {
        buf := make([]byte, 1024)
        n, err := r.Body.Read(buf)
        if err != nil && err != io.EOF {
          log.Fatal("super bad err when reading body: ", err)
        }
        originalErr := err
        fmt.Printf("Read %d bytes from httpbin.org\n", n)
        fmt.Printf("Data: %s\n", string(buf[:n]))
        fullData = append(fullData, buf[:n]...)
        if n == 0 {
          sha256Sum := sha256.Sum256(fullData)
          trailers := headers.NewHeaders()
          trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", sha256Sum))
          trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullData)))
          n, err = w.WriteChunkedBodyDone(trailers)
          if err != nil {
            log.Fatal("super bad err when chunking: ", err)
          }
          return 
        }
        n, err = w.WriteChunkedBody(buf[:n])
        if err != nil {
          log.Fatal("super bad err when reading body: ", err)
        }
        if originalErr == io.EOF {
          sha256Sum := sha256.Sum256(fullData)
          trailers := headers.NewHeaders()
          trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", sha256Sum))
          trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullData)))
          n, err = w.WriteChunkedBodyDone(trailers)
          if err != nil {
            log.Fatal("super bad err when chunking: ", err)
          }
          return 
        }
      }
    }
  }

	server, err := server.Serve(port, newH)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}