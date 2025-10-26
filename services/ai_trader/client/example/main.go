package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client"
)

func main() {
	// Example usage of the AI trader client
	nodeURL := "http://localhost:26657"

	// Create the main client
	aiClient, err := client.NewClient(nodeURL)
	if err != nil {
		log.Fatalf("Failed to create AI trader client: %v", err)
	}
	defer aiClient.Close()

	// Example 1: Get price data
	fmt.Println("=== Price Data Example ===")
	ctx := context.Background()

	price, err := aiClient.GetPriceData(ctx, "BTC")
	if err != nil {
		fmt.Printf("Error getting BTC price: %v\n", err)
	} else {
		fmt.Printf("BTC Price: %s %s (Source: %s, Confidence: %.2f)\n",
			price.Value, price.Symbol, price.Source, price.Confidence)
	}

	// Example 2: Get multiple prices
	fmt.Println("\n=== Multiple Prices Example ===")
	prices, err := aiClient.GetMultiplePriceData(ctx, []string{"BTC", "ETH", "OSMO"})
	if err != nil {
		fmt.Printf("Error getting multiple prices: %v\n", err)
	} else {
		for symbol, price := range prices {
			fmt.Printf("%s: %s (Source: %s, Confidence: %.2f)\n",
				symbol, price.Value, price.Source, price.Confidence)
		}
	}

	// Example 3: Validate price data
	fmt.Println("\n=== Price Validation Example ===")
	if price != nil {
		err := aiClient.ValidatePriceData(price)
		if err != nil {
			fmt.Printf("Price validation failed: %v\n", err)
		} else {
			fmt.Println("Price validation passed!")
		}

		// Check if price is stale
		isStale := aiClient.IsPriceStale(price, 30*time.Minute)
		fmt.Printf("Price is stale (older than 30min): %v\n", isStale)
	}

	// Example 4: Execute trading decision
	fmt.Println("\n=== Trading Decision Example ===")
	decision := &client.TradingDecision{
		Symbol:     "BTC",
		Action:     "buy",
		Amount:     "1000000", // 1M NDOLLAR
		Price:      "50000.00",
		Reason:     "Price drop detected",
		Confidence: 0.8,
	}

	grantee := "cosmos1qk93t4j0yyzgqgt6k5qf8deh8fq6smpn3ntu3x"
	granter := "cosmos1p9qh4ldfd6n0qehujsal4k7g0e37kel90rc4ts"

	result, err := aiClient.ExecuteTradingDecision(ctx, decision, grantee, granter)
	if err != nil {
		fmt.Printf("Error executing trading decision: %v\n", err)
	} else {
		fmt.Printf("Trading decision executed successfully: %v\n", result.Success)
	}

	// Example 5: Health check
	fmt.Println("\n=== Health Check Example ===")
	err = aiClient.HealthCheck(ctx)
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
	} else {
		fmt.Println("All clients are healthy!")
	}

	fmt.Println("\n=== Example completed ===")
}
