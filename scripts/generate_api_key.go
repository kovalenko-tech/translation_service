package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func main() {
	// Generate 32 bytes (256 bits) of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		fmt.Printf("Error generating random bytes: %v\n", err)
		return
	}

	// Convert to hex string
	apiKey := hex.EncodeToString(bytes)

	fmt.Printf("Generated API Key: %s\n", apiKey)
	fmt.Printf("Add this to your .env file as: API_KEY=%s\n", apiKey)
}
