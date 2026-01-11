package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"http_server/internal/headers"
	"http_server/internal/request"
	"http_server/internal/response"
	"http_server/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 8080

func main() {
	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	log.Printf("Received request: %s %s\n", req.Line.Method, req.Line.RequestTarget)

	if strings.HasPrefix(req.Line.RequestTarget, "/httpbin") {
		handleProxyRequest(w, req)
		return
	}

	if req.Line.Method != "GET" {
		writeErrorHTML(w, response.MethodNotAllowed, "Method Not Allowed", "")
		return
	}

	switch req.Line.RequestTarget {
	case "/yourproblem":
		writeErrorHTML(w, response.BadRequest, "Bad Request", "Your request honestly kinda sucked.")
	case "/myproblem":
		writeErrorHTML(w, response.InternalServerError, "Internal Server Error", "Okay, you know what? This one is on me.")
	case "/video":
		handleVideoRequest(w)
	default:
		writeSuccessHTML(w)
	}
}

func handleProxyRequest(w *response.Writer, req *request.Request) {
	path := strings.TrimPrefix(req.Line.RequestTarget, "/httpbin")
	targetURL := "https://httpbin.org" + path

	log.Printf("Proxying request to: %s\n", targetURL)

	resp, err := http.Get(targetURL)
	if err != nil {
		log.Printf("Error making proxy request: %v\n", err)
		writeErrorHTML(w, response.InternalServerError, "Internal Server Error", "Failed to proxy request")
		return
	}
	defer resp.Body.Close()

	if err := w.WriteStatusLine(response.StatusCode(resp.StatusCode)); err != nil {
		log.Printf("Error writing status line: %v\n", err)
		return
	}

	h := response.GetDefaultHeaders(0)
	delete(h, "content-length")
	h.Set("transfer-encoding", "chunked")
	h.Set("trailer", "X-Content-SHA256, X-Content-Length")

	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		h.Set("content-type", contentType)
	}

	if err := w.WriteHeaders(h); err != nil {
		log.Printf("Error writing headers: %v\n", err)
		return
	}

	buffer := make([]byte, 1024)
	var fullBody []byte
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			log.Printf("Read %d bytes from httpbin\n", n)
			fullBody = append(fullBody, buffer[:n]...)
			if _, writeErr := w.WriteChunkedBody(buffer[:n]); writeErr != nil {
				log.Printf("Error writing chunked body: %v\n", writeErr)
				return
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading response body: %v\n", err)
			return
		}
	}

	if _, err := w.WriteChunkedBodyDone(); err != nil {
		log.Printf("Error writing final chunk: %v\n", err)
		return
	}

	hash := sha256.Sum256(fullBody)
	hashHex := hex.EncodeToString(hash[:])

	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", hashHex)
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))

	if err := w.WriteTrailers(trailers); err != nil {
		log.Printf("Error writing trailers: %v\n", err)
	}
}

func handleVideoRequest(w *response.Writer) {
	videoData, err := os.ReadFile("assets/video.mp4")
	if err != nil {
		log.Printf("Error reading video file: %v\n", err)
		writeErrorHTML(w, response.InternalServerError, "Internal Server Error", "Failed to read video file")
		return
	}

	if err := w.WriteStatusLine(response.OK); err != nil {
		log.Printf("Error writing status line: %v\n", err)
		return
	}

	h := response.GetDefaultHeaders(len(videoData))
	h.Set("content-type", "video/mp4")

	if err := w.WriteHeaders(h); err != nil {
		log.Printf("Error writing headers: %v\n", err)
		return
	}

	if _, err := w.WriteBody(videoData); err != nil {
		log.Printf("Error writing body: %v\n", err)
	}
}

func writeErrorHTML(w *response.Writer, statusCode response.StatusCode, title string, message string) {
	body := fmt.Sprintf(`<html>
	  <head>
		<title>%d %s</title>
	  </head>
	  <body>
		<h1>%s</h1>
		<p>%s</p>
	  </body>
	</html>`, statusCode, title, title, message)

	h := response.GetDefaultHeaders(len(body))
	h.Set("content-type", "text/html")

	if err := w.WriteStatusLine(statusCode); err != nil {
		log.Println("Error writing status line:", err)
		return
	}

	if err := w.WriteHeaders(h); err != nil {
		log.Println("Error writing headers:", err)
		return
	}

	if _, err := w.WriteBody([]byte(body)); err != nil {
		log.Println("Error writing body:", err)
	}
}

func writeSuccessHTML(w *response.Writer) {
	body := `<html>
	  <head>
		<title>200 OK</title>
	  </head>
	  <body>
		<h1>Success!</h1>
		<p>Your request was an absolute banger.</p>
	  </body>
	</html>`

	h := response.GetDefaultHeaders(len(body))
	h.Set("content-type", "text/html")

	if err := w.WriteStatusLine(response.OK); err != nil {
		log.Println("Error writing status line:", err)
		return
	}

	if err := w.WriteHeaders(h); err != nil {
		log.Println("Error writing headers:", err)
		return
	}

	if _, err := w.WriteBody([]byte(body)); err != nil {
		log.Println("Error writing body:", err)
	}
}
