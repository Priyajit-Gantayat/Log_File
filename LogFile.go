package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sync"
)

type LogEntry struct {
	Original string `json:"original"`
	Pattern  string `json:"pattern,omitempty"`
}

func main() {
	// Path to the log file
	logFile := "sample.log"

	// Open the log file
	file, err := os.Open(logFile)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer file.Close()

	// Channel for log lines and results
	lines := make(chan string, 1000)
	results := make(chan LogEntry, 1000)

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Start workers to process lines
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go processLines(lines, results, &wg)
	}

	// Start a goroutine to read the file and send lines to the channel
	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()

	// Wait for workers to finish and close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and process results
	duplicates := make(map[string]bool)
	var processed []LogEntry

	for result := range results {
		// Check for duplicates
		if !duplicates[result.Original] {
			processed = append(processed, result)
			duplicates[result.Original] = true
		}
	}

	// Convert results to JSON
	jsonData, err := json.MarshalIndent(processed, "", "  ")
	if err != nil {
		fmt.Printf("Error converting results to JSON: %v\n", err)
		return
	}

	// Save JSON to a file
	err = os.WriteFile("output.json", jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON to file: %v\n", err)
		return
	}

	fmt.Println("Processing completed. Results saved to output.json")
}

// processLines processes log lines, performs pattern matching, and sends results to the channel
func processLines(lines <-chan string, results chan<- LogEntry, wg *sync.WaitGroup) {
	defer wg.Done()

	// Define your regex patterns
	patterns := []string{
		`ERROR`,   // Match lines containing "ERROR"
		`WARNING`, // Match lines containing "WARNING"
		`DEBUG`,   // Match lines containing "DEBUG"
	}

	for line := range lines {
		for _, pattern := range patterns {
			matched, _ := regexp.MatchString(pattern, line)
			if matched {
				results <- LogEntry{
					Original: line,
					Pattern:  pattern,
				}
				break
			}
		}
	}
}
