package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	rootDirectory := flag.String("r", "C:\\", "root directory to begin walking for stats")
	maxDepth := flag.Int("d", 1, "depth for how many subfolders to walkthrough via recursion")
	topN := flag.Int("n", 10, "number of largest files to track")
	flag.Parse()

	if *maxDepth < 0 {
		fmt.Printf("Scanning Directory: %s (depth: ∞)\n", *rootDirectory)
	} else {
		fmt.Printf("Scanning Directory: %s (depth: %d)\n", *rootDirectory, *maxDepth)
	}

	start := time.Now()
	ConcurrentWalk(*rootDirectory, *maxDepth, *topN)
	fmt.Printf("Scan Duration: %v\n", time.Since(start))
}
