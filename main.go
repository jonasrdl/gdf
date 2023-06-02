package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sync"
)

var (
	baseURL      = "url"
	wordlistPath = "wordlist.txt"
	workers      = 10
)

type Result struct {
	URL string
}

func main() {
	wordlist, err := ReadWordlist(wordlistPath)
	if err != nil {
		fmt.Printf("Error reading wordlist file: %s\n", err)
		return
	}

	totalPaths := len(wordlist)

	var wg sync.WaitGroup

	results := make(chan Result)

	progress := make(chan int)

	go PrintResults(results)

	go UpdateProgress(progress, totalPaths)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go Worker(i, wordlist, &wg, results, progress)
	}

	wg.Wait()

	close(results)
	close(progress)
}

func Worker(workerID int, wordlist []string, wg *sync.WaitGroup, results chan<- Result, progress chan<- int) {
	defer wg.Done()

	for index, path := range wordlist {
		url := baseURL + path

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error making request to %s: %s\n", url, err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			results <- Result{URL: url}
		}

		resp.Body.Close()

		progress <- index + 1
	}
}

func PrintResults(results <-chan Result) {
	foundFiles := make(map[string]bool)

	for result := range results {
		if !foundFiles[result.URL] {
			fmt.Printf("Found: %s\n", result.URL)
			foundFiles[result.URL] = true
		}
	}
}

func UpdateProgress(progress <-chan int, total int) {
	for current := range progress {
		fmt.Printf("\rProgress: (%d/%d)", current, total)
	}

	fmt.Println()
}

func ReadWordlist(path string) ([]string, error) {
	wordlist := []string{}

	file, err := os.Open(path)
	if err != nil {
		return wordlist, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		wordlist = append(wordlist, scanner.Text())
	}

	if scanner.Err() != nil {
		return wordlist, scanner.Err()
	}

	return wordlist, nil
}
