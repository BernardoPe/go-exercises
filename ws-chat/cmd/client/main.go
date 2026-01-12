package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"ws-chat/internal/client"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	serverAddr := flag.String("server", "ws://localhost:8080/ws", "WebSocket server address")
	flag.Parse()
	ctx := context.Background()
	chatClient, err := client.NewClient(*serverAddr)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	if err := chatClient.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to %s: %v", *serverAddr, err)
	}
	defer chatClient.Close()

	model := client.NewTUIModel(chatClient)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
