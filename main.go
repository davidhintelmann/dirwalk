package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("can not resolve working directory: %v\n", err)
	}

	rootDirectory := flag.String("r", dir, "root directory to begin walking for stats")
	maxDepth := flag.Int("d", 0, "depth for how many subfolders to walkthrough via recursion (default: 0  no recursion)")
	topN := flag.Int("n", 10, "number of largest files to track and output")
	flag.Parse()

	if *maxDepth < 0 {
		fmt.Printf("Scanning Directory: %s (depth: ∞)\n", *rootDirectory)
	} else {
		fmt.Printf("Scanning Directory: %s (depth: %d)\n", *rootDirectory, *maxDepth)
	}

	start := time.Now()
	results := ConcurrentWalk(*rootDirectory, *maxDepth, *topN)
	results.PrintOutput()
	fmt.Printf("Scan Duration: %v\n", time.Since(start))
}
