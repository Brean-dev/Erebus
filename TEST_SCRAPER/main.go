package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	url := "http://localhost:8080"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("error creating request: %v", err)
	}

	// --- optional headers (simulate a browser or scraper) ---
	req.Header.Set("User-Agent", "Go-http-client/1.1")
	req.Header.Set("Accept", "*/*")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response: %v", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Body:\n%s\n", body)
}
