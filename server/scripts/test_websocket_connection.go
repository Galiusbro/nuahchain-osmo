package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
)

// Simple test to verify WebSocket client works
func main() {
	wsURL := "ws://localhost:26657/websocket"
	if len(os.Args) > 1 {
		wsURL = os.Args[1]
	}

	fmt.Printf("Testing WebSocket connection to: %s\n", wsURL)

	client := blockchain.NewWebSocketClient(
		wsURL,
		blockchain.WebSocketConfig{
			ReconnectInterval: 5 * time.Second,
			Timeout:           30 * time.Second,
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to connect
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	fmt.Println("✅ WebSocket connected successfully")

	// Test subscription to new blocks
	testTxHash := "test1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	fmt.Printf("\nTesting subscription to transaction: %s\n", testTxHash)

	subCtx, subCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer subCancel()

	sub, err := client.Subscribe(subCtx, testTxHash)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	fmt.Println("✅ Subscription created")

	// Wait for event or timeout
	select {
	case event := <-sub.Events:
		fmt.Printf("✅ Received event: %+v\n", event)
	case <-time.After(5 * time.Second):
		fmt.Println("⚠️  No event received (timeout - this is expected for non-existent tx)")
	}

	// Check connection status
	if client.IsConnected() {
		fmt.Println("✅ WebSocket still connected")
	} else {
		fmt.Println("❌ WebSocket disconnected")
	}

	fmt.Printf("Active subscriptions: %d\n", client.GetSubscriptionCount())

	if err := client.GetLastError(); err != nil {
		fmt.Printf("Last error: %v\n", err)
	}

	fmt.Println("\n✅ WebSocket test completed")
}
