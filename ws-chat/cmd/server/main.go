package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"ws-chat/internal/chat"
	"ws-chat/internal/server"
)

func main() {
	repo := chat.NewInMemoryRoomRepository()
	service := chat.NewService(repo)
	wsServer := server.NewServer(service)

	http.HandleFunc("/ws", wsServer.HandleWebSocket())

	port := ":8080"
	s := &http.Server{
		Addr: port,
	}

	go func() {
		log.Printf("WebSocket server starting on %s", port)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("ListenAndServe error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wsServer.Close()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
