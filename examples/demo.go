package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// Simple demo application that demonstrates the key features of MockOho
// This script makes requests to both mocked and proxied endpoints
// and shows how to toggle endpoints and change responses

func main() {
	// Start MockOho server in the background
	fmt.Println("Starting MockOho server...")
	cmd := exec.Command("../mockoho", "server")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Failed to start MockOho server: %v\n", err)
		return
	}
	defer cmd.Process.Kill()

	// Wait for the server to start
	fmt.Println("Waiting for server to start...")
	time.Sleep(2 * time.Second)

	// Make a request to a mocked endpoint
	fmt.Println("\n1. Making request to mocked endpoint...")
	resp, err := http.Get("http://localhost:3000/api/hello")
	if err != nil {
		fmt.Printf("Failed to make request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read and print the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}
	fmt.Printf("Response from mocked endpoint: %s\n", body)

	// Parse the response to pretty print it
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err == nil {
		prettyJSON, _ := json.MarshalIndent(data, "", "  ")
		fmt.Printf("Pretty response:\n%s\n", prettyJSON)
	}

	// Make a request to a proxied endpoint (will be proxied if not mocked)
	fmt.Println("\n2. Making request to proxied endpoint...")
	resp, err = http.Get("http://localhost:3000/api/users")
	if err != nil {
		fmt.Printf("Failed to make request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read and print the response
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}
	fmt.Printf("Response from proxied endpoint: %s\n", body)

	// Demonstrate how to modify the mock configuration
	fmt.Println("\n3. Modifying mock configuration...")
	fmt.Println("To modify the mock configuration, you would normally use the MockOho CLI:")
	fmt.Println("  - Press 't' to toggle an endpoint active/inactive")
	fmt.Println("  - Press 'r' to cycle through available responses")
	fmt.Println("  - Press 'o' to open the configuration file in your editor")

	// Manually modify a mock configuration file to demonstrate
	fmt.Println("\n4. Manually modifying a mock configuration file...")
	examplePath := "../mocks/example.json"
	
	// Read the current configuration
	configData, err := os.ReadFile(examplePath)
	if err != nil {
		fmt.Printf("Failed to read configuration: %v\n", err)
		return
	}

	// Parse the configuration
	var exampleConfig map[string]interface{}
	err = json.Unmarshal(configData, &exampleConfig)
	if err != nil {
		fmt.Printf("Failed to parse configuration: %v\n", err)
		return
	}

	// Modify the configuration
	endpoints := exampleConfig["endpoints"].([]interface{})
	endpoint := endpoints[0].(map[string]interface{})
	
	// Toggle the endpoint active state
	endpoint["active"] = !(endpoint["active"].(bool))
	
	// Save the modified configuration
	modifiedConfig, err := json.MarshalIndent(exampleConfig, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal configuration: %v\n", err)
		return
	}
	
	err = os.WriteFile(examplePath, modifiedConfig, 0644)
	if err != nil {
		fmt.Printf("Failed to write configuration: %v\n", err)
		return
	}
	
	fmt.Printf("Modified configuration saved to %s\n", examplePath)
	fmt.Println("Endpoint active state toggled")

	// Wait a moment for the server to reload the configuration
	time.Sleep(1 * time.Second)

	// Make another request to see the effect of the change
	fmt.Println("\n5. Making request to the modified endpoint...")
	resp, err = http.Get("http://localhost:3000/api/hello")
	if err != nil {
		fmt.Printf("Failed to make request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read and print the response
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}
	fmt.Printf("Response after modification: %s\n", body)

	fmt.Println("\nDemo completed successfully!")
}